package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-querystring/query"
	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/logx"
	"github.com/natansdj/go_scrape/models"
	"net/http"
	"net/url"
	"strconv"
)

func scrapeFundHandler(cfg config.ConfYaml) gin.HandlerFunc {
	return func(c *gin.Context) {

		baseUri := cfg.Source.BaseURI

		form := url.Values{}
		form.Add("firstopen", "yes")
		form.Add("aumlowervalue", "500")
		form.Add("aumlowercheck", "yes")
		form.Add("aumbetweenlowvalue", "500")
		form.Add("aumbetweenhighvalue", "2000")
		form.Add("aumbetweencheck", "yes")
		form.Add("aumgreatervalue", "2000")
		form.Add("aumgreatercheck", "yes")
		form.Add("availibility", "available")
		form.Add("fundtype", "mm,fi,balance,equity")
		form.Add("hiloselect", "1yr")
		form.Add("performancetype", "nav")
		form.Add("fundnonsyariah", "yes")
		form.Add("fundsyariah", "yes")
		form.Add("etfnonsyariah", "yes")
		form.Add("etfsyariah", "yes")

		req, _ := RequestInit(cfg, "GET", "source_json_for_favorite.php", nil, form)
		body, err := RequestDo(req)
		if err != nil {
			logx.LogError.Error(err.Error())
			panic(err)
		}

		var i gin.H
		err = json.NewDecoder(NewJSONReader(body)).Decode(&i)
		if err != nil {
			logx.LogError.Error(err.Error())
			panic(err.Error())
		} else {
			logx.LogAccess.Info(fmt.Sprintf("\nType : %T", i["aaData"]))

			//list of funds
			if aaData, ok := i["aaData"].([]interface{}); ok {
				logx.LogAccess.Info(fmt.Sprintf("Len : %v", len(aaData)))

				//Process each fund
				var funds []models.Funds
				var managers []models.Manager
				for k, aDtRaw := range aaData {
					logx.LogAccess.Info(fmt.Sprintf("FUND : %v, %T", aaData[k], aaData[k]))
					if aDtVal, ok2 := aDtRaw.([]interface{}); ok2 {
						var f models.Funds

						//Process each fund attribute
						for l := range aDtVal {
							var v string
							if v, ok = aDtVal[l].(string); !ok {
								continue
							}

							switch l {
							case 0: //id
								f.FundId, _ = strconv.Atoi(v)
							case 1: //id
							case 2: //name
								f.FundName = v
							case 3: //manager
								f.MiName = v
							case 4: //fund_type
								f.FundType = v
							case 5: //last_nav
								f.LastNAV, _ = strconv.ParseFloat(v, 64)
							case 6: //1d
								f.D1, _ = strconv.ParseFloat(v, 64)
							case 7: //3d
								f.D3, _ = strconv.ParseFloat(v, 64)
							case 8: //1m
								f.M1, _ = strconv.ParseFloat(v, 64)
							case 9: //3m
								f.M3, _ = strconv.ParseFloat(v, 64)
							case 10: //6m
								f.M6, _ = strconv.ParseFloat(v, 64)
							case 11: //9m
								f.M9, _ = strconv.ParseFloat(v, 64)
							case 12: //ytd
								f.YTD, _ = strconv.ParseFloat(v, 64)
							case 13: //1yr
								f.Y1, _ = strconv.ParseFloat(v, 64)
							case 14: //3yr
								f.Y3, _ = strconv.ParseFloat(v, 64)
							case 15: //5yr
								f.Y5, _ = strconv.ParseFloat(v, 64)
							case 16: //hi-lo
								f.HiLo, _ = strconv.ParseFloat(v, 64)
							case 17: //sharpe
								f.Sharpe, _ = strconv.ParseFloat(v, 64)
							case 18: //drawdown
								f.DrawDown, _ = strconv.ParseFloat(v, 64)
							case 19: //dd_periode
								f.DdPeriode, _ = strconv.Atoi(v)
							case 20: //
							case 21: //
							case 22: //hist_risk
								f.HistRisk, _ = strconv.ParseFloat(v, 64)
							case 23: //aum
								f.AUM, _ = strconv.ParseFloat(v, 64)
							case 24: //morningstar
								f.Morningstar, _ = strconv.ParseFloat(v, 64)
							case 25: //code
								f.MiCode = v
							case 26: //active
								f.Active, _ = strconv.Atoi(v)
							case 27: //
							case 28: //
							case 29: //
							case 30: //risk
								f.Risk = v
							case 31: //type
								f.Type = v
							}
						}

						//manager
						if f.MiCode != "" {
							mi := models.Manager{MiCode: f.MiCode, MiName: f.MiName}
							managers = append(managers, mi)
						}

						//funds
						if f.FundId > 0 && f.FundName != "" {
							funds = append(funds, f)
						}
					}
				}

				//Batch insert/update
				if len(managers) > 0 {
					for _, v := range managers {
						//Create if not exists
						err = models.ManagerCreateIfNotExists(&v)
					}
				}

				if len(funds) > 0 {
					for _, v := range funds {
						//Create or update fund
						err = models.FundCreateOrUpdate(&v)
					}
				}

			}
			fmt.Println("")
		}

		urlStr := req.URL.String()
		c.JSON(http.StatusOK, gin.H{
			"source":  urlStr,
			"baseUri": baseUri,
			"result":  i,
		})
	}
}

type StCompChart struct {
	Type      string `form:"type" json:"type" url:"type"`
	Jenis     string `form:"jenis" json:"jenis" url:"jenis" `
	FundId    string `form:"fundid" json:"fundid" url:"fundid" `
	StartDate string `form:"startdate" json:"startdate" url:"startdate"`
	EndDate   string `form:"enddate" json:"enddate" url:"enddate"`
}

type RetCompChart struct {
	All        string `json:"all"`
	EndData    string `json:"enddata"`
	OneDay     string `json:"oneday"`
	OneMonth   string `json:"onemonth"`
	ThreeMonth string `json:"threemonth"`
	SixMonth   string `json:"sixmonth"`
	OneYear    string `json:"oneyear"`
	ThreeYear  string `json:"threeyear"`
	Ytd        string `json:"ytd"`
}

func scrapeNavHandler(cfg config.ConfYaml) gin.HandlerFunc {
	return func(c *gin.Context) {
		var a StCompChart
		err := c.Bind(&a)
		if err != nil {
			logx.LogError.Error(err.Error())
			panic(err)
		}

		//Fetch datatanggal
		if a.Type == "" {
			a.Type = "getdatatanggal"
		}
		if a.Jenis == "" {
			a.Jenis = "nav"
		}
		if a.FundId == "" {
			a.FundId = "4"
		}
		urlValues, _ := query.Values(a)

		req, _ := RequestInit(cfg, "GET", "comparison_chart_json.php", nil, urlValues)
		body, err := RequestDo(req)
		if err != nil {
			logx.LogError.Error(err.Error())
			panic(err)
		}

		var i gin.H
		err = json.NewDecoder(NewJSONReader(body)).Decode(&i)

		//Fetch Nav
		req2, i2 := ipFetchNav(cfg, a)

		//RESULT
		c.JSON(http.StatusOK, gin.H{
			"source1": req.URL.String(),
			"source2": req2.URL.String(),
			"baseUri": cfg.Source.BaseURI,
			"form":    a,
			"result1": i,
			"result2": i2,
		})
	}
}

func ipFetchNav(cfg config.ConfYaml, a StCompChart) (*http.Request, interface{}) {
	//Fetch Nav
	b := StCompChart{
		Type:      "popupnav",
		StartDate: "MjUgTWF5IDIwMDU=",
		EndDate:   "MTYgSnVsIDIwMjE=",
		FundId:    a.FundId,
	}

	urlValues, _ := query.Values(b)

	req2, _ := RequestInit(cfg, "GET", "comparison_chart_json.php", nil, urlValues)
	body2, err2 := RequestDo(req2)
	if err2 != nil {
		logx.LogError.Error(err2.Error())
		panic(err2)
	}

	fmt.Println(string(body2))

	var i2 gin.H

	jr := new(JSONReader)
	jr.Reader = bytes.NewReader(body2)
	err2 = json.NewDecoder(jr).Decode(&i2)

	return req2, i2
}
