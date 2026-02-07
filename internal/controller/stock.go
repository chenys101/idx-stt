package controller

import (
	"encoding/json"
	_ "log"
	_ "strconv"
	"stt/internal/model"
	"stt/internal/pkg/app"
	"stt/internal/pkg/database"
	"stt/internal/pkg/encrypt"
	"stt/internal/pkg/stock"
	"stt/internal/pkg/strutil"

	"github.com/gin-gonic/gin"
)

type StockController struct{}

func (sc *StockController) GetStockData(c *gin.Context) {
	list := c.Query("list")
	if list == "" {
		code := c.Query("code")
		// 简单验证 code 是否为空
		if code == "" {
			code = "default"
		}
		var stockMonitor model.StockMonitor
		err := database.DB.First(&stockMonitor, "code = ?", code).Error
		if err != nil {
			if err.Error() == "record not found" {
				app.AbortWithError(c, 404, "StockMonitor record not found")
			} else {
				app.AbortWithError(c, 500, err.Error())
			}
			return
		}
		list = stockMonitor.MonitorValue
	}

	// 调用 stock.GetStockData 函数获取股票数据
	stockData, err := stock.GetStockData(list)
	if err != nil {
		app.AbortWithError(c, 500, err.Error())
		return
	}
	// stockData 转为 StockBase
	stockBaseEn := make([]model.StockBaseEn, len(*stockData))
	//log.Printf("data-1: %v\n", stockData)
	for i, v := range *stockData {
		nameArr := []rune(v.Name)
		nameNoise := strutil.AddNoise(string(nameArr[0]))
		//log.Printf("cpc: %v\n",nameNoise + "----   " + strconv.FormatFloat(v.ChangePercent, 'f', 2, 64))
		stockBaseEn[i] = model.StockBaseEn{
			Name: nameNoise + strutil.FloatNumToChinese(v.ChangePrice) + "#" + strutil.FloatNumToChinese(v.ChangePercent),
			//log.Printf("cpc: %v\n",nameNoise + "----" + v.ChangePercent)
			//Cp: 	strutil.FloatNumToChinese(v.ChangePercent),
			//Cpr:   	strutil.FloatNumToChinese(v.ChangePrice),
		}
	}
	// 将 stockBase 序列化为 JSON 字符串
	stockBaseJSON, err := json.Marshal(stockBaseEn)
	if err != nil {
		app.AbortWithError(c, 500, err.Error())
		return
	}
	// stockBase 改为加密 aes
	encryptor, err := encrypt.NewEncryptor(encrypt.AES256, []byte("k7z49c2m5p8b3q6r"))
	if err != nil {
		app.AbortWithError(c, 500, err.Error())
		return
	}
	encryptedData, err := encryptor.Encrypt(stockBaseJSON)
	if err != nil {
		app.AbortWithError(c, 500, err.Error())
		return
	}
	// 返回加密后的数据
	app.Success(c, 200, encryptedData)
}

func (sc *StockController) GetStockDataNotEncrypt(c *gin.Context) {
	list := c.Query("list")
	if list == "" {
		code := c.Query("code")
		// 简单验证 code 是否为空
		if code == "" {
			code = "default"
		}
		var stockMonitor model.StockMonitor
		err := database.DB.First(&stockMonitor, "code = ?", code).Error
		if err != nil {
			if err.Error() == "record not found" {
				app.AbortWithError(c, 404, "StockMonitor record not found")
			} else {
				app.AbortWithError(c, 500, err.Error())
			}
			return
		}
		list = stockMonitor.MonitorValue
	}

	// 调用 stock.GetStockData 函数获取股票数据
	stockData, err := stock.GetStockData(list)
	if err != nil {
		app.AbortWithError(c, 500, err.Error())
		return
	}
	// stockData 转为 StockBase
	stockBase := make([]model.StockBase, len(*stockData))
	for i, v := range *stockData {
		nameArr := []rune(v.Name)
		nameNoise := strutil.AddNoise(string(nameArr[0]))
		stockBase[i] = model.StockBase{
			Name:          nameNoise,
			ChangePercent: v.ChangePercent,
			ChangePrice:   v.ChangePrice,
		}
	}
	app.Success(c, 200, stockBase)
}

func (sc *StockController) CreateStockMonitor(c *gin.Context) {
	var input struct {
		Code         string `json:"code" binding:"required"`
		MonitorValue string `json:"monitorValue" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		app.AbortWithError(c, 400, err.Error())
		return
	}
	stockMonitor := model.StockMonitor{
		Code:         input.Code,
		MonitorValue: input.MonitorValue,
	}
	if result := database.DB.Create(&stockMonitor); result.Error != nil {
		app.AbortWithError(c, 500, result.Error.Error())
		return
	}
	app.Success(c, 201, stockMonitor)
}

func (sc *StockController) UpdateStockMonitor(c *gin.Context) {
	var input struct {
		Code         string `json:"code" binding:"required"`
		MonitorValue string `json:"monitorValue" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		app.AbortWithError(c, 400, err.Error())
		return
	}

	var stockMonitor model.StockMonitor
	err := database.DB.First(&stockMonitor, "code = ?", input.Code).Error
	if err != nil {
		if err.Error() == "record not found" {
			app.AbortWithError(c, 404, "StockMonitor record not found")
		} else {
			app.AbortWithError(c, 500, err.Error())
		}
		return
	}

	stockMonitor.MonitorValue = input.MonitorValue
	result := database.DB.Save(&stockMonitor)
	if result.Error != nil {
		app.AbortWithError(c, 500, result.Error.Error())
		return
	}

	app.Success(c, 200, stockMonitor)
}

func (sc *StockController) GetAllStockMonitors(c *gin.Context) {
	var stockMonitors []model.StockMonitor
	err := database.DB.Find(&stockMonitors).Error
	if err != nil {
		app.AbortWithError(c, 500, err.Error())
		return
	}

	app.Success(c, 200, stockMonitors)
}
