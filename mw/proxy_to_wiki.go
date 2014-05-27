package mw

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"../mwutils"
	"../nlog"
)

const DefaultWikiaName = "www"
const WikiaNameQueryParam = "wikianame"
const WikiaLangQueryParam = "wikialang"

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func defaultWikiBaseUrl(wikiaName, wikiaLang string) *url.URL {
	u := new(url.URL)
	baseHost := "wikia.com"
	u.Scheme = "http"
	if wikiaLang != "en" && wikiaLang != "" {
		u.Host = wikiaLang + "." + wikiaName + "." + baseHost
	} else {
		u.Host = wikiaName + "." + baseHost
	}
	return u
}

func WikiProxy(wikiBaseUrl func(wikiaName, wikiaLang string) *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{Director: defaultWikiProxyDirector()}
}

func defaultWikiProxyDirector() func(req *http.Request) {
	return func(req *http.Request) {
		con, ok := mwutils.MapperGetOk(req)
		if ok == false {
			panic("couldn't get connection object from global pool")
		}
		target := con.Metadata.TargetWikiaUrl

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		nlog.Debugf("default wiki request: %+v", req)

		if target.RawQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
		}
	}
}

func WikiaDesignationQueryParser(con *mwutils.Connection) bool {
	query := con.Request.URL.Query()
	wikiaName := query.Get(WikiaNameQueryParam)
	if wikiaName != "" {
		query.Del(WikiaNameQueryParam)
	}
	wikiaLang := query.Get(WikiaLangQueryParam)
	if wikiaLang != "" {
		query.Del(WikiaLangQueryParam)
	}
	con.Request.URL.RawQuery = query.Encode()

	con.Metadata.Wikia = &mwutils.Wikia{wikiaName, wikiaLang}
	return true
}

func DefaultTargetWikiaURL(con *mwutils.Connection) bool {
	u := new(url.URL)
	baseHost := "wikia.com"
	u.Scheme = "http"
	wikiaName := con.Metadata.Wikia.Name
	wikiaLang := con.Metadata.Wikia.Lang

	if wikiaName == "" {
		wikiaName = "www"
	}
	if wikiaLang != "en" && wikiaLang != "" {
		u.Host = wikiaLang + "." + wikiaName + "." + baseHost
	} else {
		u.Host = wikiaName + "." + baseHost
	}
	con.Metadata.TargetWikiaUrl = u
	return true
}

func TWikiProxy(con *mwutils.Connection) bool {
	return false
}

func DefaultWikiProxy() func(rw http.ResponseWriter, req *http.Request) {
	return WikiProxy(defaultWikiBaseUrl).ServeHTTP
}
