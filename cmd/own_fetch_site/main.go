package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"rudy_gc/internal/config"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/fetchjavbus"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/fetchsukebei"

	"github.com/zeromicro/go-zero/core/conf"
)

var (
	gidPattern   = regexp.MustCompile(`gid\s*=\s*(\d+)`)
	imgPattern   = regexp.MustCompile(`img\s*=\s*['"]([^'"]+)['"]`)
	ucPattern    = regexp.MustCompile(`uc\s*=\s*(\d+)`)
	floorPattern = regexp.MustCompile(`floor\s*=\s*(\d+)`)

	configFile = flag.String("f", "cmd/api/config.yaml", "the config file")
	movieJavID = flag.String("movie_jav_id", "", "movie jav id")
	movieCode  = flag.String("movie_code", "", "movie code")
	mode       = flag.String("mode", "both", "javbus|sukebei|both")
	enqueue    = flag.Bool("enqueue", true, "ensure fetch tasks before run")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if strings.TrimSpace(*movieJavID) == "" || strings.TrimSpace(*movieCode) == "" {
		panic("movie_jav_id and movie_code are required")
	}

	d, err := dep.New(c)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	siteSvc := fetchsite.NewService(d)
	releaseDate := int64(0)
	if movieRow, findErr := d.MovieModel.FindOneByJavId(ctx, strings.TrimSpace(*movieJavID)); findErr == nil {
		releaseDate = movieRow.ReleasingDate
	}
	fmt.Printf("config.fetcher.proxy=%s\n", c.Fetcher.Proxy)
	if row, err := d.FetchSiteModel.FindOneBySiteCode(ctx, fetchsite.FetchSiteCodeJavbus); err == nil {
		fmt.Printf("db.javbus.proxy=%s\n", row.Proxy)
	} else {
		fmt.Printf("db.javbus.proxy.err=%v\n", err)
	}
	if row, err := d.FetchSiteModel.FindOneBySiteCode(ctx, fetchsite.FetchSiteCodeSukebei); err == nil {
		fmt.Printf("db.sukebei.proxy=%s\n", row.Proxy)
	} else {
		fmt.Printf("db.sukebei.proxy.err=%v\n", err)
	}
	fmt.Printf("proxy.javbus=%s\n", d.FetchSites[fetchsite.FetchSiteCodeJavbus].Proxy)
	fmt.Printf("proxy.sukebei=%s\n", d.FetchSites[fetchsite.FetchSiteCodeSukebei].Proxy)
	if resp, err := d.Fetcher.GetBySite(ctx, fetchsite.FetchSiteCodeJavbus, "https://www.javbus.com/"+strings.ToLower(strings.TrimSpace(*movieCode))); err == nil {
		body := string(resp.Body)
		if len(body) > 600 {
			body = body[:600]
		}
		fmt.Printf("prefetch.javbus.status=%d\n", resp.Status)
		fmt.Printf("prefetch.javbus.body=%q\n", body)
		gid := findMatch(gidPattern, string(resp.Body))
		img := findMatch(imgPattern, string(resp.Body))
		uc := findMatch(ucPattern, string(resp.Body))
		floor := findMatch(floorPattern, string(resp.Body))
		if uc == "" {
			uc = "0"
		}
		if floor == "" {
			floor = "1"
		}
		if gid != "" && img != "" {
			values := url.Values{}
			values.Set("gid", gid)
			values.Set("lang", "zh")
			values.Set("img", img)
			values.Set("uc", uc)
			values.Set("floor", floor)
			ajaxURL := "https://www.javbus.com/ajax/uncledatoolsbyajax.php?" + values.Encode()
			if ajaxResp, ajaxErr := d.Fetcher.GetBySiteWithOptions(ctx, fetchsite.FetchSiteCodeJavbus, ajaxURL, fetchsite.RequestOptions{
				Referer: "https://www.javbus.com/" + strings.ToLower(strings.TrimSpace(*movieCode)),
				Headers: map[string]string{
					"X-Requested-With": "XMLHttpRequest",
				},
			}); ajaxErr == nil {
				ajaxBody := string(ajaxResp.Body)
				if len(ajaxBody) > 600 {
					ajaxBody = ajaxBody[:600]
				}
				fmt.Printf("prefetch.javbus.ajax.url=%s\n", ajaxURL)
				fmt.Printf("prefetch.javbus.ajax.status=%d\n", ajaxResp.Status)
				fmt.Printf("prefetch.javbus.ajax.body=%q\n", ajaxBody)
			} else {
				fmt.Printf("prefetch.javbus.ajax.err=%v\n", ajaxErr)
			}
		}
	} else {
		fmt.Printf("prefetch.javbus.err=%v\n", err)
	}
	if *enqueue {
		fmt.Println("ensure fetch tasks...")
		if err := siteSvc.EnsureFetchTasksForMovie(ctx, strings.TrimSpace(*movieJavID), strings.TrimSpace(*movieCode), releaseDate); err != nil {
			panic(err)
		}
	}

	currentMode := strings.ToLower(strings.TrimSpace(*mode))
	if currentMode == "" {
		currentMode = "both"
	}

	if currentMode == "javbus" || currentMode == "both" {
		fmt.Println("run javbus...")
		javbusSvc := fetchjavbus.NewService(d)
		rows, err := javbusSvc.FetchMovieMagnets(ctx, strings.TrimSpace(*movieJavID), strings.TrimSpace(*movieCode))
		if err != nil {
			panic(err)
		}
		fmt.Printf("javbus rows=%d\n", len(rows))
	}

	if currentMode == "both" {
		fmt.Println("sleep before sukebei...")
		if err := siteSvc.SleepRequest(ctx, fetchsite.FetchSiteCodeSukebei); err != nil {
			panic(err)
		}
	}

	if currentMode == "sukebei" || currentMode == "both" {
		fmt.Println("run sukebei...")
		sukebeiSvc := fetchsukebei.NewService(d)
		rows, err := sukebeiSvc.FetchMovieTorrents(ctx, strings.TrimSpace(*movieJavID), strings.TrimSpace(*movieCode))
		if err != nil {
			panic(err)
		}
		fmt.Printf("sukebei rows=%d\n", len(rows))
	}
}

func findMatch(pattern *regexp.Regexp, text string) string {
	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
