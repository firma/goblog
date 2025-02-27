package anqicms

import (
	stdContext "context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/skratchdot/open-golang/open"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/crond"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/middleware"
	"kandaoni.com/anqicms/provider"
	"kandaoni.com/anqicms/route"
	"kandaoni.com/anqicms/tags"
	"kandaoni.com/anqicms/view"
	"log"
	"os"
	"strings"
	"time"
)

type Bootstrap struct {
	Application *iris.Application
	Port        int
	LoggerLevel string
	viewEngine  *view.DjangoEngine
}

func New(port int, loggerLevel string) *Bootstrap {
	var bootstrap Bootstrap
	bootstrap.Port = port
	bootstrap.LoggerLevel = loggerLevel

	return &bootstrap
}

func (bootstrap *Bootstrap) loadGlobalMiddleware() {
	bootstrap.Application.Use(middleware.NewRecover())
	bootstrap.Application.Use(middleware.Cors)
	bootstrap.Application.Options("{path:path}", middleware.Cors)
}

func (bootstrap *Bootstrap) Serve() {
	//自动迁移表
	if provider.GetDefaultDB() != nil {
		provider.InitWebsites()
	}

	//开始计划任务
	crond.Crond()

	go bootstrap.Start()
	go func() {
		time.Sleep(1 * time.Second)
		currentSite := provider.CurrentSite(nil)
		link := fmt.Sprintf("http://127.0.0.1:%d", bootstrap.Port)
		if currentSite != nil && currentSite.System.BaseUrl != "" {
			if strings.Contains(currentSite.System.BaseUrl, "127.0.0.1") {
				currentSite.System.BaseUrl = link
			}
			link = currentSite.System.BaseUrl
		}
		err := open.Run(link)
		if err != nil {
			log.Println("请手动在浏览器输入访问地址：", link)
		}
	}()
	// 伪静态规则和模板更改变化
	for {
		select {
		case restart := <-config.RestartChan:
			if restart {
				fmt.Println("监听到路由更改")
				_ = bootstrap.Shutdown()
				log.Println("进程结束，开始重启")
				// 重启
				_ = provider.Restart()
			} else {
				// reload template
				fmt.Println("重载模板")
				bootstrap.viewEngine.Load()
			}
		}
	}
}

func (bootstrap *Bootstrap) Start() {
	bootstrap.Application = iris.New()
	bootstrap.Application.Logger().SetLevel(bootstrap.LoggerLevel)
	bootstrap.loadGlobalMiddleware()
	route.Register(bootstrap.Application)

	pugEngine := view.Django(".html")
	// 开发模式下动态加载
	if config.Server.Server.Env == "development" {
		pugEngine.Reload(true)
	}

	pugEngine.AddFunc("stampToDate", TimestampToDate)

	_ = pugEngine.RegisterTag("tdk", tags.TagTdkParser)
	_ = pugEngine.RegisterTag("system", tags.TagSystemParser)
	_ = pugEngine.RegisterTag("contact", tags.TagContactParser)
	_ = pugEngine.RegisterTag("navList", tags.TagNavListParser)
	_ = pugEngine.RegisterTag("categoryList", tags.TagCategoryListParser)
	_ = pugEngine.RegisterTag("categoryDetail", tags.TagCategoryDetailParser)
	_ = pugEngine.RegisterTag("archiveDetail", tags.TagArchiveDetailParser)
	_ = pugEngine.RegisterTag("pageList", tags.TagPageListParser)
	_ = pugEngine.RegisterTag("pageDetail", tags.TagPageDetailParser)
	_ = pugEngine.RegisterTag("prevArchive", tags.TagPrevArchiveParser)
	_ = pugEngine.RegisterTag("nextArchive", tags.TagNextArchiveParser)
	_ = pugEngine.RegisterTag("archiveList", tags.TagArchiveListParser)
	_ = pugEngine.RegisterTag("breadcrumb", tags.TagBreadcrumbParser)
	_ = pugEngine.RegisterTag("pagination", tags.TagPaginationParser)
	_ = pugEngine.RegisterTag("linkList", tags.TagLinkListParser)
	_ = pugEngine.RegisterTag("commentList", tags.TagCommentListParser)
	_ = pugEngine.RegisterTag("guestbook", tags.TagGuestbookParser)
	_ = pugEngine.RegisterTag("archiveParams", tags.TagArchiveParamsParser)
	_ = pugEngine.RegisterTag("tagList", tags.TagTagListParser)
	_ = pugEngine.RegisterTag("tagDetail", tags.TagTagDetailParser)
	_ = pugEngine.RegisterTag("tagDataList", tags.TagTagDataListParser)
	_ = pugEngine.RegisterTag("archiveFilters", tags.TagArchiveFiltersParser)
	_ = pugEngine.RegisterTag("userDetail", tags.TagUserDetailParser)
	_ = pugEngine.RegisterTag("userGroupDetail", tags.TagUserGroupDetailParser)
	_ = pugEngine.RegisterTag("bannerList", tags.TagBannerListParser)

	bootstrap.viewEngine = pugEngine
	// 模板在最后加载，避免因为模板而导致程序无法运行
	bootstrap.Application.RegisterView(pugEngine)

	err := bootstrap.Application.Run(
		iris.Addr(fmt.Sprintf(":%d", bootstrap.Port)),
		iris.WithRemoteAddrHeader("X-Real-IP"),
		iris.WithRemoteAddrHeader("X-Forwarded-For"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithoutBodyConsumptionOnUnmarshal,
		iris.WithoutPathCorrection,
	)

	if err != nil {
		log.Println(err.Error())
		library.DebugLog(config.ExecPath, "error.log", time.Now().Format("2006-01-02 15:04:05"), "启动服务出错", err.Error())
		os.Exit(0)
	}
}

func TimestampToDate(in int64, layout string) string {
	t := time.Unix(in, 0)
	return t.Format(layout)
}

func (bootstrap *Bootstrap) Shutdown() error {
	bootstrap.Application.Shutdown(stdContext.Background())

	return nil
}
