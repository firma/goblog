package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/provider"
	"kandaoni.com/anqicms/response"
	"strings"
)

func CategoryPage(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	cacheFile, ok := currentSite.LoadCachedHtml(ctx)
	if ok {
		ctx.ServeFile(cacheFile)
		return
	}
	currentPage := ctx.Values().GetIntDefault("page", 1)
	categoryId := ctx.Params().GetUintDefault("id", 0)
	catId := ctx.Params().GetUintDefault("catid", 0)
	if catId > 0 {
		categoryId = catId
	}
	urlToken := ctx.Params().GetString("filename")
	multiCatNames := ctx.Params().GetString("multicatname")
	if multiCatNames != "" {
		chunkCatNames := strings.Split(multiCatNames, "/")
		urlToken = chunkCatNames[len(chunkCatNames)-1]
		var prev *model.Category
		for _, catName := range chunkCatNames {
			tmpCat := currentSite.GetCategoryFromCacheByToken(catName)
			if tmpCat == nil || (prev != nil && tmpCat.ParentId != prev.Id) {
				NotFound(ctx)
				return
			}
			prev = tmpCat
		}
	}

	var category *model.Category
	var err error
	if urlToken != "" {
		//优先使用urlToken
		category, err = currentSite.GetCategoryByUrlToken(urlToken)
	} else {
		category, err = currentSite.GetCategoryById(categoryId)
	}
	if err != nil {
		NotFound(ctx)
		return
	}

	//修正，如果这里读到的的page，则跳到page中
	if category.Type == config.CategoryTypePage {
		ctx.StatusCode(301)
		ctx.Redirect(currentSite.GetUrl("page", category, 0))
		return
	}

	module := currentSite.GetModuleFromCache(category.ModuleId)
	if module == nil {
		ctx.StatusCode(404)
		ShowMessage(ctx, currentSite.Lang("未定义模型"), nil)
		return
	}

	if webInfo, ok := ctx.Value("webInfo").(*response.WebInfo); ok {
		webInfo.Title = category.Title
		if category.SeoTitle != "" {
			webInfo.Title = category.SeoTitle
		}
		if currentPage > 1 {
			webInfo.Title += " - " + fmt.Sprintf(currentSite.Lang("第%d页"), currentPage)
		}
		webInfo.Keywords = category.Keywords
		webInfo.Description = category.Description
		webInfo.NavBar = category.Id
		webInfo.PageName = "archiveList"
		webInfo.CanonicalUrl = currentSite.GetUrl("category", category, currentPage)
		ctx.ViewData("webInfo", webInfo)
	}

	ctx.ViewData("category", category)

	tplName := module.TableName + "/list.html"
	tplName2 := module.TableName + "_list.html"
	if ViewExists(ctx, tplName2) {
		tplName = tplName2
	}
	//模板优先级：1、设置的template；2、存在分类id为名称的模板；3、继承的上级模板；4、默认模板，如果发现上一级不继承，则不需要处理
	if category.Template != "" {
		tplName = category.Template
	} else if ViewExists(ctx, fmt.Sprintf("%s/list-%d.html", module.TableName, category.Id)) {
		tplName = fmt.Sprintf("%s/list-%d.html", module.TableName, category.Id)
	} else {
		categoryTemplate := currentSite.GetCategoryTemplate(category)
		if categoryTemplate != nil && len(categoryTemplate.Template) > 0 {
			tplName = categoryTemplate.Template
		}
	}
	if !strings.HasSuffix(tplName, ".html") {
		tplName += ".html"
	}
	recorder := ctx.Recorder()
	err = ctx.View(GetViewPath(ctx, tplName))
	if err != nil {
		ctx.Values().Set("message", err.Error())
	} else {
		if currentSite.PluginHtmlCache.Open && currentSite.PluginHtmlCache.IndexCache > 0 {
			_ = currentSite.CacheHtmlData(ctx.RequestPath(false), ctx.Request().URL.RawQuery, ctx.IsMobile(), recorder.Body())
		}
	}
}

func SearchPage(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	cacheFile, ok := currentSite.LoadCachedHtml(ctx)
	if ok {
		ctx.ServeFile(cacheFile)
		return
	}
	q := strings.TrimSpace(ctx.URLParam("q"))
	moduleToken := ctx.Params().GetString("module")
	var module *model.Module
	if len(moduleToken) > 0 {
		module = currentSite.GetModuleFromCacheByToken(moduleToken)
		if module == nil {
			ctx.StatusCode(404)
			ShowMessage(ctx, currentSite.Lang("未定义模型"), nil)
			return
		}
		ctx.ViewData("module", module)
	}
	if currentSite.Safe.ContentForbidden != "" {
		forbiddens := strings.Split(currentSite.Safe.ContentForbidden, "\n")
		for _, v := range forbiddens {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if strings.Contains(q, v) {
				ShowMessage(ctx, currentSite.Lang("您搜索的关键词包含有不允许的字符"), nil)
				return
			}
		}
	}

	if webInfo, ok := ctx.Value("webInfo").(*response.WebInfo); ok {
		currentPage := ctx.Values().GetIntDefault("page", 1)
		webInfo.Title = fmt.Sprintf("%s: %s", currentSite.Lang("搜索"), q)
		if module != nil {
			webInfo.Title = module.Title + webInfo.Title
		}
		if currentPage > 1 {
			webInfo.Title += " - " + fmt.Sprintf(currentSite.Lang("第%d页"), currentPage)
		}
		webInfo.PageName = "search"
		webInfo.CanonicalUrl = currentSite.GetUrl(fmt.Sprintf("/search?q=%s(&page={page})", q), nil, currentPage)
		ctx.ViewData("webInfo", webInfo)
	}

	ctx.ViewData("q", q)

	tplName := "search/index.html"
	if ViewExists(ctx, "search.html") {
		tplName = "search.html"
	}
	if module != nil {
		if ViewExists(ctx, "search/"+module.UrlToken+".html") {
			tplName = "search/" + module.UrlToken + ".html"
		} else if ViewExists(ctx, "search_"+module.UrlToken+".html") {
			tplName = "search_" + module.UrlToken + ".html"
		}
	}
	recorder := ctx.Recorder()
	err := ctx.View(GetViewPath(ctx, tplName))
	if err != nil {
		ctx.Values().Set("message", err.Error())
	} else {
		if currentSite.PluginHtmlCache.Open && currentSite.PluginHtmlCache.IndexCache > 0 {
			_ = currentSite.CacheHtmlData(ctx.RequestPath(false), ctx.Request().URL.RawQuery, ctx.IsMobile(), recorder.Body())
		}
	}
}
