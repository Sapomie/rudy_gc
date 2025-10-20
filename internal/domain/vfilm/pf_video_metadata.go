package vfilm

import (
	"encoding/json"
	"errors"
	"fmt"
	"rudy_gc/pkg/convert"
	"strconv"
	"strings"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"github.com/xfrr/goffmpeg/models"
)

type videoMeta struct {
	DataStr      string
	Width        int64
	Height       int64
	BitRate      int64
	Duration     int64
	FrameAverage float64
}

func filmMetaData(filmPath string) (*videoMeta, error) {
	dataStr, metaData, err := getMetadata(filmPath)
	if err != nil {
		return nil, err
	}
	var fr string
	vm := &videoMeta{DataStr: dataStr}
	for _, stream := range metaData.Streams {
		if stream.CodecType == "video" {
			vm.Width = int64(stream.Width)
			vm.Height = int64(stream.Height)
			fr = stream.AvgFrameRate
			break
		}
	}

	bitRate, err := strconv.Atoi(metaData.Format.BitRate)
	if err != nil {
		return nil, err
	}
	durFloat, err := strconv.ParseFloat(metaData.Format.Duration, 64)
	if err != nil {
		durFloat = 0
	}
	frame, err := getFrame(fr)
	if err != nil {
		return nil, err
	}
	vm.BitRate = int64(bitRate)
	vm.Duration = int64(durFloat)
	vm.FrameAverage = frame

	return vm, nil
}

func getMetadata(file string) (string, *models.Metadata, error) {
	dataStr, err := ffmpeg_go.Probe(file)
	if err != nil {
		return "", nil, err
	}
	metaData := new(models.Metadata)
	err = json.Unmarshal([]byte(dataStr), metaData)
	if err != nil {
		return "", nil, err
	}
	return dataStr, metaData, nil
}

func getFrame(fraction string) (float64, error) {
	parts := strings.Split(fraction, "/")
	if len(parts) != 2 {
		return 0, errors.New("输入格式不正确，应为 '分子/分母'")
	}

	s1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("分子解析错误: %v", err)
	}

	s2, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("分母解析错误: %v", err)
	}

	rate := convert.FloatTo(s1 / s2).Decimal(2)

	return rate, nil
}
