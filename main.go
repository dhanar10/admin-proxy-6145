package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/elazarl/goproxy"
)

func main() {
	getLoginUserRegex := regexp.MustCompile("get_login_user")
	isLoginedRegex := regexp.MustCompile("is_logined")

	proxy := goproxy.NewProxyHttpServer()

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if !getLoginUserRegex.MatchString(ctx.Req.URL.String()) {
				return resp
			}

			log.Println(ctx.Req.URL.String())

			originalBody := new(strings.Builder)

			_, err := io.Copy(originalBody, resp.Body)

			if err != nil {
				return goproxy.NewResponse(ctx.Req, goproxy.ContentTypeText, http.StatusBadGateway, err.Error())
			}

			var bodyMap map[string]interface{}

			json.Unmarshal([]byte(originalBody.String()), &bodyMap)

			bodyMap["login_user"] = "1" // Patch get_login_user response

			patchedBody, err := json.Marshal(bodyMap)

			if err != nil {
				return goproxy.NewResponse(ctx.Req, goproxy.ContentTypeText, http.StatusBadGateway, err.Error())
			}

			return goproxy.NewResponse(ctx.Req, "application/json", http.StatusOK, string(patchedBody))
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if !isLoginedRegex.MatchString(ctx.Req.URL.String()) {
				return resp
			}

			log.Println(ctx.Req.URL.String())

			return goproxy.NewResponse(ctx.Req, "application/json", http.StatusOK, `{"result":"1","user":"1"}`) // Overwrite is_logined response
		})

	log.Fatal(http.ListenAndServe(":51080", proxy))
}
