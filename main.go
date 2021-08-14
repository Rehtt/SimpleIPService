package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lionsoul2014/ip2region/binding/golang/ip2region"
	"github.com/rehtt/go-map"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"net/http"
	"strings"
)
var flag_download =false

const db = "ip2region.db"
func main() {
	if _,err:=ioutil.ReadFile(db);err!=nil{
		downloadDB()
	}
	c:=cron.New()
	c.AddFunc("0 1 * * *", func() {
		downloadDB()
	})
	r := gin.Default()
	r.GET("/", func(context *gin.Context) {
		context.Writer.WriteString(getRemoteIp(context))
	})
	r.GET("/:ip", func(context *gin.Context) {
		remoteIp := context.Param("ip")
		if remoteIp == "ip" {
			remoteIp = getRemoteIp(context)
		}
		info, err := getIpInfo(remoteIp)
		if err != nil {
			context.Writer.WriteString(err.Error())
			return
		}
		outJson := context.Query("json")
		if outJson == "true" {
			j, _ := json.Marshal(info.Pull2Map())
			context.Writer.Write(j)
			return
		}
		var str strings.Builder
		keys, values := info.Pull2List()
		for i := range keys {
			str.WriteString(keys[i])
			str.WriteString(" : ")
			str.WriteString(fmt.Sprintf("%v", values[i]))
			str.WriteString("\n")
		}
		context.Writer.WriteString(str.String())

	})

	r.Run("0.0.0.0:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getRemoteIp(context *gin.Context) string {
	return context.ClientIP()
}
func getIpInfo(ip string) (*go_map.Map, error) {
	region, err := ip2region.New(db)
	defer region.Close()
	if err != nil {
		if !flag_download{
			downloadDB()
		}else{
			return nil, fmt.Errorf("系统更新中，请稍后")
		}

	}
	info, err := region.BtreeSearch(ip)
	if err != nil {
		return &go_map.Map{}, err
	}
	data := go_map.New()
	data.Set("ip", ip)
	data.Set("国家", info.Country)
	data.Set("省份", info.Province)
	data.Set("区域", info.Region)
	data.Set("城市", info.City)
	data.Set("城市ID", info.CityId)
	data.Set("ISP", info.ISP)

	return data, nil
}

func downloadDB() {
	flag_download=true
	fmt.Println("start")
	resp, err := http.Get("https://github.com/lionsoul2014/ip2region/raw/master/data/ip2region.db")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	err = ioutil.WriteFile(db, data, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("stop")
	flag_download=false
}
