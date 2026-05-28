(function () {
  const sidebarHTML = `
<div class="vertical-menu">
    <div data-simplebar class="h-100">
        <div id="sidebar-menu">
            <ul class="metismenu list-unstyled" id="side-menu">
                <li class="menu-title">Main</li>

                <li>
                    <a href="javascript: void(0);" class="has-arrow waves-effect">
                        <i class="mdi mdi-book-open-outline"></i><span>MovieCard</span>
                    </a>
                    <ul class="sub-menu" aria-expanded="false">
                        <li><a href="/cardstoday">Today</a></li>
                        <li><a href="/cardsmediamowned">MediaOwned</a></li>
                        <li><a href="/cardshasrank">HasRank</a></li>
                        <li><a href="/moviecarddayrank">DayRank</a></li>
                        <li><a href="/moviecardperiodrank">PeriodRank</a></li>
                        <li><a href="/cardsneeddownload">NeedDownload</a></li>
                        <li><a href="/cardsrandom">Random</a></li>
                    </ul>
                </li>

                <li>
                    <a href="javascript: void(0);" class="has-arrow waves-effect">
                        <i class="mdi mdi-chart-box-outline"></i><span>MovieAgg</span>
                    </a>
                    <ul class="sub-menu" aria-expanded="false">
                        <li><a href="/movie-agg-all/release">Agg-All-Release</a></li>
                        <li><a href="/movie-agg-all/release-bucket-list">Agg-All-Release-Bucket</a></li>
                        <li><a href="/w-media-agg/bucket-list">Agg-WMedia-Bucket</a></li>
                        <li><a href="/w-media-agg/birth">Agg-WMedia-Birth</a></li>
                        <li><a href="/triggers/agg">AggTriggers</a></li>
                        <li><a href="/wdir">WDirectory</a></li>
                    </ul>
                </li>

                <li>
                    <a href="javascript: void(0);" class="has-arrow waves-effect">
                        <i class="mdi mdi-format-list-bulleted"></i><span>List</span>
                    </a>
                    <ul class="sub-menu" aria-expanded="false">
                        <li><a href="/medias">Medias</a></li>
                        <li><a href="/w-agg-events">WAggEvents</a></li>
                        <li><a href="/torrent-albums">TorrentAlbums</a></li>
                        <li><a href="/movie-albums">MovieAlbums</a></li>
                        <li><a href="/d-seeds">DSeeds</a></li>
                        <li><a href="/casts">Casts</a></li>
                        <li><a href="/e-items">EItems</a></li>
                    </ul>
                </li>

                <li>
                    <a href="javascript: void(0);" class="has-arrow waves-effect">
                        <i class="mdi mdi-star-outline"></i><span>SC</span>
                    </a>
                    <ul class="sub-menu" aria-expanded="false">
                        <li><a href="/sc-events">ScEvents</a></li>
                        <li><a href="/sc-pick-smart-media">ScPickSmartMedia</a></li>
                        <li><a href="/triggers/sc">ScPreview</a></li>
                        <li><a href="/triggers/sc-media">ScMediaMove</a></li>
                    </ul>
                </li>

                <li>
                    <a href="javascript: void(0);" class="has-arrow waves-effect">
                        <i class="mdi mdi-clipboard-text-outline"></i><span>抓取任务</span>
                    </a>
                    <ul class="sub-menu" aria-expanded="false">
                        <li><a href="/crawl-records">CrawlRecords</a></li>
                        <li><a href="/triggers/dailybest">DailyBest</a></li>
                        <li><a href="/triggers/seeds">Seeds</a></li>
                        <li><a href="/triggers/post-process">PostProcess</a></li>
                        <li><a href="/triggers/backfill">Backfill</a></li>
                        <li><a href="/crawler/detail-loop">DetailLoop</a></li>
                        <li><a href="/crawler/tasks">CrawlerTasks</a></li>
                    </ul>
                </li>

                <li>
                    <a href="javascript: void(0);" class="has-arrow waves-effect">
                        <i class="mdi mdi-web"></i><span>FetchSite操作</span>
                    </a>
                    <ul class="sub-menu" aria-expanded="false">
                        <li><a href="/triggers/fetch-site">FetchSite</a></li>
                        <li><a href="/triggers/fetch-sehuatang">FetchSehuatang</a></li>
                        <li><a href="/fetch-site-javbus-list">JavbusList</a></li>
                        <li><a href="/fetch-site-sukebei-list">SukebeiList</a></li>
                        <li><a href="/fetch-site-sehuatang-list">SehuatangList</a></li>
                    </ul>
                </li>

                <li>
                    <a href="javascript: void(0);" class="has-arrow waves-effect">
                        <i class="mdi mdi-folder-cog-outline"></i><span>Media操作</span>
                    </a>
                    <ul class="sub-menu" aria-expanded="false">
                        <li><a href="/triggers/media">Media</a></li>
                        <li><a href="/triggers/media-rescan">MediaRescan</a></li>
                        <li><a href="/triggers/media-rollback">MediaRollback</a></li>
                    </ul>
                </li>
            </ul>
        </div>
    </div>
</div>`;

  const root = document.getElementById("sidebar-root");
  if (root) {
    root.outerHTML = sidebarHTML;
  }
}());
