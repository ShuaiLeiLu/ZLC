package main

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"time"
	//"github.com/labstack/echo/v4/middleware"
)

func main() {
	//实例化echo对象。
	e := echo.New()

	//连接数据库
	var rsp *redis.Pool
	rsp = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     10000,
		MaxActive:   10000,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "1.117.150.173:6379", redis.DialPassword("lushuailei"), redis.DialDatabase(0))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	//rs, _ := redis.Dial("tcp", "127.0.0.1:6379", redis.DialPassword("ddoyyds"), redis.DialDatabase(0))
	fmt.Printf("OK!")

	e.GET("/bean", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "bean", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/city", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "city", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/ddfactory", func(c echo.Context) error {
		rs := rsp.Get()
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "ddfactory", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/farm", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "farm", "20"))

		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/health", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "health", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/jxfactory", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "jxfactory", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/pet", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "pet", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/sgmh", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "sgmh", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.POST("/upload/cfd", func(c echo.Context) error {
		rs := rsp.Get()
		returntext := ""
		ptpin := c.QueryParam("ptpin")
		code := c.QueryParam("code")
		result, _ := rs.Do("SISMEMBER", "pin", ptpin)
		account := fmt.Sprintf("%v", result)
		accountint, _ := strconv.Atoi(account)
		if accountint == 1 {
			if len(code)%32 == 0 {
				result, _ := rs.Do("SISMEMBER", "cfdpin", ptpin)
				amount := fmt.Sprintf("%v", result)
				amountint, _ := strconv.Atoi(amount)
				if amountint == 0 {
					result, _ := rs.Do("SCARD", "cfd")
					amount := fmt.Sprintf("%v", result)
					amountint, _ := strconv.Atoi(amount)
					if amountint < 3000 {
						_, _ = rs.Do("SADD", "cfd", code)
						_, _ = rs.Do("SADD", "cfdpin", ptpin)
						returntext = "OK"
					} else {
						returntext = "full"
					}
				} else {
					returntext = "exist"
				}
			} else {
				returntext = "error"
			}
		} else {
			returntext = "not in whitelist"
		}
		rs.Close()
		return c.String(http.StatusOK, returntext)
	})
	e.GET("/cfd", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "cfd", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/carnivalcity", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "carnivalcity", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.GET("/jxmc", func(c echo.Context) error {
		//控制器函数直接返回一个字符串，http响应状态为http.StatusOK，就是200状态。
		rs := rsp.Get()
		result, _ := redis.Strings(rs.Do("SRANDMEMBER", "jxmc", "20"))
		rs.Close()
		jsondata, _ := json.Marshal(result)
		return c.String(http.StatusOK, "{\"data\":"+string(jsondata)+",\"code\":200}")
	})
	e.POST("/upload/jxmc", func(c echo.Context) error {
		rs := rsp.Get()
		returntext := ""
		ptpin := c.QueryParam("ptpin")
		code := c.QueryParam("code")
		result, _ := rs.Do("SISMEMBER", "pin", ptpin)
		account := fmt.Sprintf("%v", result)
		accountint, _ := strconv.Atoi(account)
		if accountint == 1 {
			if code != "" {
				result, _ := rs.Do("SISMEMBER", "jxmcpin", ptpin)
				amount := fmt.Sprintf("%v", result)
				amountint, _ := strconv.Atoi(amount)
				if amountint == 0 {
					result, _ := rs.Do("SCARD", "jxmc")
					amount := fmt.Sprintf("%v", result)
					amountint, _ := strconv.Atoi(amount)
					if amountint < 2000 {
						_, _ = rs.Do("SADD", "jxmc", code)
						_, _ = rs.Do("SADD", "jxmcpin", ptpin)
						returntext = "OK"
					} else {
						returntext = "full"
					}
				} else {
					returntext = "exist"
				}
			} else {
				returntext = "error"
			}
		} else {
			returntext = "not in whitelist"
		}
		rs.Close()
		return c.String(http.StatusOK, returntext)
	})
	e.GET("/test1", func(c echo.Context) error {
		fmt.Printf(c.QueryParam("id"))
		return c.String(http.StatusOK, c.QueryParam("id"))
	})

	e.Start(":80")
}
