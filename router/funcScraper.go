package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/natansdj/go_scrape/logx"
	"github.com/natansdj/go_scrape/models"
	"github.com/natansdj/go_scrape/utils"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	GoProcessCount = 3
)

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

func scrapeFundsHandler() gin.HandlerFunc {
	cfg := models.CFG
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
			logx.LogAccess.Infof("\nType : %T", i["aaData"])

			//list of funds
			if aaData, ok := i["aaData"].([]interface{}); ok {
				logx.LogAccess.Infof("Len : %v", len(aaData))

				//Process each fund
				var funds []models.Funds
				var managers []models.Manager
				for k, aDtRaw := range aaData {
					logx.LogAccess.Infof("FUND : %v, %T", aaData[k], aaData[k])
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

func scrapeNavsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalFunds, totalProcessed int
		var ids []int

		if f := c.Query("id"); f != "" {
			fsplit := strings.Split(f, ",")
			for k := range fsplit {
				if id, err := strconv.Atoi(fsplit[k]); err == nil {
					ids = append(ids, id)
				}
			}
			logx.LogAccess.Infof("Query Ids : %v", ids)
		}

		var update int
		updateQ := c.Query("update")
		if v, err := strconv.Atoi(updateQ); err == nil {
			update = v
		}

		funds, err := models.FundGetAll(ids...)
		if err == nil {
			totalFunds = len(funds)
			logx.LogAccess.Infof("TOTAL : %v", len(funds))
			for _, v := range funds {
				fmt.Println()
				logx.LogAccess.Infof("FUNDID : %v FundName : %v", v.FundId, v.FundName)

				var a StCompChart
				a.FundId = strconv.Itoa(v.FundId)

				//Process single nav
				_, _, _, _, totalNavs := processNav(a, update)

				if totalNavs > 0 {
					totalProcessed += 1
				}
			}
		}

		//RESULT
		c.JSON(http.StatusOK, gin.H{
			"totalFund": totalFunds,
			"processed": totalProcessed,
		})
	}
}

func scrapeNavHandler() gin.HandlerFunc {
	cfg := models.CFG
	return func(c *gin.Context) {
		var a StCompChart
		err := c.Bind(&a)
		if err != nil {
			logx.LogError.Error(err.Error())
			panic(err)
		}

		fundId := c.Param("id")
		if fundId != "" {
			logx.LogAccess.Debug("fundId : " + fundId)
			a.FundId = fundId
		}

		//Validate
		if a.FundId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fId, err2 := strconv.Atoi(a.FundId)
		if err2 == nil {
			f, err := models.FundGetById(fId)
			if err != nil || f == nil {
				errMsg := "fund not found"
				if err != nil {
					errMsg = err.Error()
				}
				c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
				return
			}
		}

		//Process single nav
		req, i, req2, _, totalNavs := processNav(a, 1)

		//RESULT
		c.JSON(http.StatusOK, gin.H{
			"source1": req.URL.String(),
			"source2": req2.URL.String(),
			"baseUri": cfg.Source.BaseURI,
			"form":    a,
			"result1": i,
			"result2": totalNavs,
		})
	}
}

func processNav(a StCompChart, update int) (req *http.Request, i map[string]interface{}, req2 *http.Request, i2 interface{}, totalNavs int64) {
	//Check current record
	fId, _ := strconv.Atoi(a.FundId)
	currCount, err := models.NavCountByFundId(fId)
	if err == nil {
		fmt.Println(currCount)
	}

	//Start
	cfg := models.CFG
	req, i = ipFetchDate(cfg, &a)
	logx.LogAccess.Infof("StartDate : %v encoded : %v EndDate : %v encoded : %v", i["all"], a.StartDate, i["oneday"], a.EndDate)

	//Fetch Nav
	req2, i2 = ipFetchNav(cfg, &a)
	if v, ok := i2.([]interface{}); ok {
		if len(v) > 0 {
			if update == 1 && currCount > 0 {
				if int(currCount) == len(v) {
					logx.LogAccess.Infof("SKIPPED...")
					return &http.Request{}, nil, &http.Request{}, nil, currCount
				} else {
					logx.LogAccess.Infof("UPDATING....", currCount, len(v))
				}
			} else {
				logx.LogAccess.Infof("INSERT OR UPDATE...")
			}

			var wg sync.WaitGroup
			totalNavs = int64(len(v))

			//Divide slice into equal parts
			middleIdx := math.Ceil(float64(totalNavs) / GoProcessCount)
			divided := utils.Chunks(v, int(middleIdx))

			//disable db debug log
			models.DbDebugUnset()

			//wait group
			start := time.Now()
			wg.Add(len(divided))

			//Process each divided parts
			for _, w := range divided {
				go insertNavIntoDb(a, w, &wg)
			}

			//Wait
			wg.Wait()

			//INFO
			logx.LogAccess.Infof("FundId : %v, LEN : %d, MID : %v, Type : %T, Process Time : %s", a.FundId, len(v), middleIdx, v, time.Since(start))
			fmt.Println()

			//re-enable db debug log
			models.DbDebugSet()
		}
	}

	return req, i, req2, i2, totalNavs
}

func insertNavIntoDb(a StCompChart, v []interface{}, wg *sync.WaitGroup) {
	//start := time.Now()
	//defer fmt.Println(fmt.Sprintf("Process Time : %s", time.Since(start)))
	defer wg.Done()

	for k := range v {
		//fmt.Println(fmt.Sprintf("%T %f", v[k], v[k]))
		if w, ok := v[k].([]interface{}); ok {
			//fmt.Println(fmt.Sprintf("%T %f", w[0], w[0]))
			//fmt.Println(fmt.Sprintf("%T %f", w[1], w[1]))

			fundId, _ := strconv.Atoi(a.FundId)
			ts := w[0].(float64)
			navValue := w[1].(float64)

			//timestamp int64
			i, err := utils.StrTo(utils.ToStr(w[0])).Int64()
			if err != nil {
				panic(err)
			}
			//Get seconds from millisecond
			if utils.RecursionCountDigits(int(i)) <= 13 {
				i = i / 1000
				ts = ts / 1000
			}

			nav := models.Navs{
				FundId:    fundId,
				Date:      time.Unix(i, 0),
				Timestamp: int(ts),
				Value:     navValue,
			}

			//Save model
			if err := models.NavCreateOrUpdate(&nav); err != nil {
				logx.LogError.Error(err.Error())
			}
		}
	}
}
