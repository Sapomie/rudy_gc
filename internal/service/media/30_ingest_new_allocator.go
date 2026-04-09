package media

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/wfoldertree"
)

var (
	yearBucketNameRe = regexp.MustCompile(`^\d{4}-\d{3}$`)
	dayBucketNameRe  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-\d{3}$`)
)

func (s *Service) allocateTargetDirectory(ctx context.Context, layout rootLayout, now time.Time) (string, int64, error) {
	return s.allocateTargetDirectoryUnder(ctx, layout.media, now)
}

func (s *Service) allocateWatchedTargetDirectory(ctx context.Context, layout rootLayout, now time.Time) (string, int64, error) {
	return s.allocateTargetDirectoryUnder(ctx, layout.watched, now)
}

func (s *Service) allocateTargetDirectoryUnder(ctx context.Context, baseDir string, now time.Time) (string, int64, error) {
	mediaRoot := filepath.Clean(strings.TrimSpace(baseDir))
	if mediaRoot == "" {
		return "", 0, fmt.Errorf("base media dir is empty")
	}
	if err := os.MkdirAll(mediaRoot, defaultFilePerm); err != nil {
		return "", 0, err
	}

	nowUnix := now.Unix()
	if _, err := wfoldertree.EnsurePathChain(ctx, s.deps.WFolderModel, consts.WFolderSourceNative, mediaRoot, nowUnix); err != nil {
		return "", 0, err
	}

	yearPath, _, err := chooseYearBucket(mediaRoot, now)
	if err != nil {
		return "", 0, err
	}
	if err := os.MkdirAll(yearPath, defaultFilePerm); err != nil {
		return "", 0, err
	}

	if _, err := wfoldertree.EnsurePathChain(ctx, s.deps.WFolderModel, consts.WFolderSourceNative, yearPath, nowUnix); err != nil {
		return "", 0, err
	}

	dayName, err := chooseDayBucket(yearPath, now)
	if err != nil {
		return "", 0, err
	}
	dayPath := filepath.Join(yearPath, dayName)
	if err := os.MkdirAll(dayPath, defaultFilePerm); err != nil {
		return "", 0, err
	}

	dayFolder, err := wfoldertree.EnsurePathChain(ctx, s.deps.WFolderModel, consts.WFolderSourceNative, dayPath, nowUnix)
	if err != nil {
		return "", 0, err
	}

	return dayPath, dayFolder.Id, nil
}

func previewTargetDirectory(layout rootLayout, now time.Time) (string, error) {
	return previewTargetDirectoryUnder(layout.media, now)
}

func previewWatchedTargetDirectory(layout rootLayout, now time.Time) (string, error) {
	return previewTargetDirectoryUnder(layout.watched, now)
}

func previewTargetDirectoryUnder(baseDir string, now time.Time) (string, error) {
	mediaRoot := filepath.Clean(strings.TrimSpace(baseDir))
	if mediaRoot == "" {
		return "", fmt.Errorf("base media dir is empty")
	}

	yearPath, yearName, err := chooseYearBucket(mediaRoot, now)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		yearName = fmt.Sprintf("%04d-001", now.Year())
		yearPath = filepath.Join(mediaRoot, yearName)
	}
	dayName, err := chooseDayBucket(yearPath, now)
	if err != nil {
		return "", err
	}
	return filepath.Join(mediaRoot, yearName, dayName), nil
}

func chooseYearBucket(mediaRoot string, now time.Time) (string, string, error) {
	yearPrefix := fmt.Sprintf("%04d-", now.Year())
	yearBuckets, err := listYearBuckets(mediaRoot, yearPrefix)
	if err != nil {
		return "", "", err
	}

	if len(yearBuckets) == 0 {
		name := yearPrefix + "001"
		return filepath.Join(mediaRoot, name), name, nil
	}

	sort.Slice(yearBuckets, func(i, j int) bool { return yearBuckets[i].seq < yearBuckets[j].seq })
	for _, bucket := range yearBuckets {
		canUseBucket, err := canAllocateInYearBucket(bucket.path)
		if err != nil {
			return "", "", err
		}
		if canUseBucket {
			return bucket.path, bucket.name, nil
		}
	}

	latest := yearBuckets[len(yearBuckets)-1]
	name := fmt.Sprintf("%s%03d", yearPrefix, latest.seq+1)
	return filepath.Join(mediaRoot, name), name, nil
}

func chooseDayBucket(yearPath string, now time.Time) (string, error) {
	bucket, _, found, err := firstAvailableDayBucket(yearPath)
	if err != nil {
		return "", err
	}
	if found {
		return bucket.name, nil
	}

	return nextDayBucketName(yearPath, now)
}

func canAllocateInYearBucket(yearPath string) (bool, error) {
	_, _, found, err := firstAvailableDayBucket(yearPath)
	if err != nil {
		return false, err
	}
	if found {
		return true, nil
	}
	buckets, err := listDayBuckets(yearPath)
	if err != nil {
		return false, err
	}
	return len(buckets) < maxLeafDirsPerYear, nil
}

func firstAvailableDayBucket(yearPath string) (bucketInfo, int, bool, error) {
	buckets, err := listDayBuckets(yearPath)
	if err != nil {
		return bucketInfo{}, 0, false, err
	}

	sort.Slice(buckets, func(i, j int) bool {
		if buckets[i].name == buckets[j].name {
			return buckets[i].seq < buckets[j].seq
		}
		return buckets[i].name < buckets[j].name
	})

	for _, bucket := range buckets {
		count, err := countVideoFiles(bucket.path)
		if err != nil {
			return bucketInfo{}, 0, false, err
		}
		if count < maxFilesPerLeafDir {
			return bucket, count, true, nil
		}
	}
	return bucketInfo{}, 0, false, nil
}

func nextDayBucketName(yearPath string, now time.Time) (string, error) {
	buckets, err := listDayBuckets(yearPath)
	if err != nil {
		return "", err
	}
	if len(buckets) >= maxLeafDirsPerYear {
		return "", fmt.Errorf("year bucket is full: %s", yearPath)
	}

	dayPrefix := now.Format("2006-01-02")
	next := 1
	for _, bucket := range buckets {
		if !strings.HasPrefix(bucket.name, dayPrefix+"-") {
			continue
		}
		if bucket.seq >= next {
			next = bucket.seq + 1
		}
	}
	return fmt.Sprintf("%s-%03d", dayPrefix, next), nil
}

type bucketInfo struct {
	name string
	path string
	seq  int
}

func listYearBuckets(mediaRoot, yearPrefix string) ([]bucketInfo, error) {
	entries, err := os.ReadDir(mediaRoot)
	if err != nil {
		return nil, err
	}

	out := make([]bucketInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, yearPrefix) || !yearBucketNameRe.MatchString(name) {
			continue
		}
		seq, err := strconv.Atoi(name[len(name)-3:])
		if err != nil {
			continue
		}
		out = append(out, bucketInfo{
			name: name,
			path: filepath.Join(mediaRoot, name),
			seq:  seq,
		})
	}
	return out, nil
}

func listDayBuckets(yearPath string) ([]bucketInfo, error) {
	entries, err := os.ReadDir(yearPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	out := make([]bucketInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !dayBucketNameRe.MatchString(name) {
			continue
		}
		seq, err := strconv.Atoi(name[len(name)-3:])
		if err != nil {
			continue
		}
		out = append(out, bucketInfo{
			name: name,
			path: filepath.Join(yearPath, name),
			seq:  seq,
		})
	}
	return out, nil
}

func countVideoFiles(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if shouldSkipIngestEntryName(entry.Name()) {
			continue
		}
		if isVideoName(entry.Name()) {
			count++
		}
	}
	return count, nil
}

func ensureUniqueFileName(dir, baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	stem := strings.TrimSuffix(baseName, ext)

	for i := 0; i < 1000; i++ {
		name := baseName
		if i > 0 {
			name = fmt.Sprintf("%s_%03d%s", stem, i, ext)
		}
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if os.IsNotExist(err) {
			return name, nil
		} else {
			return "", err
		}
	}
	return "", fmt.Errorf("too many filename conflicts in %s for %s", dir, baseName)
}
