package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	//机器人初始化
	bot, err := tgbotapi.NewBotAPI("机器人token")
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	//updates, err := bot.GetUpdatesChan(u)
	updates := bot.GetUpdatesChan(u)
	go http.ListenAndServe(":8443", nil)

	//连接数据库
	/* 数据库密码

	用户:密码@tcp(ip:3306)/数据库

	*/
	db, err := sql.Open("mysql", "数据库")

	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	//db.SetMaxOpenConns(100)
	//db.SetMaxIdleConns(100)
	//rs, err := redis.Dial("tcp", "127.0.0.1:6379", redis.DialPassword("ddoyyds"), redis.DialDatabase(0))
	//if err != nil {
	//	panic(err)
	//}
	var rsp *redis.Pool
	rsp = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     10000,
		MaxActive:   10000,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "redis地址:6379", redis.DialPassword("密码"), redis.DialDatabase(0))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}

	var helptext = "欢迎来到京东滴滴车！\n" +
		"/help 显示此帮助\n" +
		"/status 查看服务器运行状态\n" +
		"/bean 提交种豆得豆助力码\n" +
		"/ddfactory 提交东东工厂助力码\n" +
		"/farm 提交东东农场助力码\n" +
		"/health 提交健康社区助力码\n" +
		"/jxfactory 提交京喜工厂助力码\n" +
		"/pet 提交东东萌宠助力码\n" +
		"/sgmh 提交闪购盲盒助力码\n" +
		"/bind 绑定京东账号\n" +
		"/bindlist 查看你的绑定列表\n" +
		"/unbind 解绑京东账号\n" +
		"/check 查看你的上车情况\n" +
		"/checkcode 查看某个助力码是否在数据库中\n" +
		"/total 查看当前上车码子总数\n" +
		"/redisclear （管理员）清空助力池\n\n" +
		"目前每种码子最多支持提交五个，如果超过五个会导致不为你添加进数据库，记住了，是一个都不加哦！\n\n" +
		"提交助力码格式：/bean code1&code2&code3&code4&code5\n" +
		"查询码子是否在数据库格式：/checkcode (bean|ddfactory|farm|health|jxfactory|pet|sgmh|cfd) yourcode\n"

	for update := range updates {

		var isGroup, codetext string
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() || update.Message.Chat.IsChannel() {
			continue
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.ReplyToMessageID = update.Message.MessageID
			if update.Message.IsCommand() {
				rs := rsp.Get()

				switch update.Message.Command() {
				case "start":
					err := db.QueryRow("SELECT isGroup FROM user where tgid=?", update.Message.Chat.ID).Scan(&isGroup)
					if err == sql.ErrNoRows {
						_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
						msg.Text = "已将您添加到数据库中！"
						bot.Send(msg)
					}
					msg.Text = fmt.Sprintf("你好，%s %s\n你的用户ID是%d\n\n%s", update.Message.Chat.FirstName, update.Message.Chat.LastName, update.Message.Chat.ID, helptext)
				case "help":
					if update.Message.Chat.IsGroup() {
						msg.Text = "请私聊我使用！"
						bot.Send(msg)
						continue
					}
					err := db.QueryRow("SELECT isGroup FROM user where tgid=?", update.Message.Chat.ID).Scan(&isGroup)
					if err == sql.ErrNoRows {
						_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
						bot.Send(msg)
					}
					msg.Text = helptext
				case "status":
					msg.Text = "服务器CPU占用率" + strconv.FormatFloat(GetCpuPercent(), 'G', -1, 32) + "%\n服务器内存占用率" + strconv.FormatFloat(GetMemPercent(), 'G', -1, 32) + "%"
				case "wiki":
					msg.ParseMode = "html"
					msg.Text = "<a href=\"https://www.example.com\">暂时还没有做WIKI~</a>"
				case "carnivalcity":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT carnivalcity FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 2 {
									result, _ := rs.Do("SCARD", "carnivalcity")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 2000 {
										_, _ = db.Exec("UPDATE user SET carnivalcity=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if len(codes[i]) != 0 {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "carnivalcity", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/carnivalcity code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/carnivalcity code1&code2&code3"
					}
				case "bean":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT bean FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 5 {
									result, _ := rs.Do("SCARD", "bean")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 12000 {
										_, _ = db.Exec("UPDATE user SET bean=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if len(codes[i])%13 == 0 {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "bean", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/bean code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/bean code1&code2&code3"
					}
				case "ddfactory":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT ddfactory FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 5 {
									result, _ := rs.Do("SCARD", "ddfactory")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 10000 {
										_, _ = db.Exec("UPDATE user SET ddfactory=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if (strings.HasPrefix(codes[i], "P0") || strings.HasPrefix(codes[i], "T0")) && strings.HasSuffix(codes[i], "RrbA") {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "ddfactory", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/ddfactory code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/ddfactory code1&code2&code3"
					}
				case "farm":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT farm FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 5 {
									result, _ := rs.Do("SCARD", "farm")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 12000 {
										_, _ = db.Exec("UPDATE user SET farm=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if len(codes[i]) == 32 {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "farm", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/farm code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/farm code1&code2&code3"
					}
				case "health":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT health FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 5 {
									result, _ := rs.Do("SCARD", "health")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 12000 {
										_, _ = db.Exec("UPDATE user SET health=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if (strings.HasPrefix(codes[i], "P0") || strings.HasPrefix(codes[i], "T0")) && strings.HasSuffix(codes[i], "RrbA") {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "health", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/health code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/health code1&code2&code3"
					}
				case "jxfactory":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT jxfactory FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 5 {
									result, _ := rs.Do("SCARD", "jxfactory")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 12000 {
										_, _ = db.Exec("UPDATE user SET jxfactory=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if len(codes[i]) == 24 || len(codes[i]) == 44 {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "jxfactory", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/jxfactory code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/jxfactory code1&code2&code3"
					}
				case "pet":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT pet FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 5 {
									result, _ := rs.Do("SCARD", "pet")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 12000 {
										_, _ = db.Exec("UPDATE user SET pet=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if strings.HasPrefix(codes[i], "MT") {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "pet", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/pet code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/pet code1&code2&code3"
					}
				case "sgmh":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT sgmh FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 5 {
									result, _ := rs.Do("SCARD", "sgmh")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 12000 {
										_, _ = db.Exec("UPDATE user SET sgmh=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											if strings.HasPrefix(codes[i], "P04z54XCjVWnYaS5u") || (strings.HasPrefix(codes[i], "T0") && strings.HasSuffix(codes[i], "rbA")) {
												if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
													msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
													bot.Send(msg)
												} else {
													_, _ = rs.Do("SADD", "sgmh", codes[i])
												}
											} else {
												msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/sgmh code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/sgmh code1&code2&code3"
					}
				case "city":
					codetext = update.Message.Text
					codetextlist := strings.Split(codetext, " ")
					if len(codetextlist) == 2 {
						codetextrow := codetextlist[1]
						codes := strings.Split(codetextrow, "&")
						newcodelen := len(codes)
						if newcodelen >= 1 {
							var oldcount int
							err := db.QueryRow("SELECT city FROM user where tgid=?", update.Message.Chat.ID).Scan(&oldcount)
							if err == sql.ErrNoRows {
								_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
								msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
							} else {
								if newcodelen+oldcount <= 2 {
									result, _ := rs.Do("SCARD", "city")
									amount := fmt.Sprintf("%v", result)
									amountint, _ := strconv.Atoi(amount)
									if amountint < 3000 {
										_, _ = db.Exec("UPDATE user SET city=? WHERE tgid=?", newcodelen+oldcount, update.Message.Chat.ID)
										for i := 0; i < newcodelen; i++ {
											//if 1{
											if strings.Contains(codes[i], "'") || strings.Contains(codes[i], "\"") || strings.Contains(codes[i], "@") || strings.Contains(codes[i], "\n") || strings.Contains(codes[i], "“") {
												msg.Text = fmt.Sprintf("%v\n存在非法字符，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
												bot.Send(msg)
											} else {
												_, _ = rs.Do("SADD", "city", codes[i])
											}
											//} else {
											//	msg.Text = fmt.Sprintf("%v\n格式错误，该码子上车失败，恭喜你浪费一次提交机会", codes[i])
											//	bot.Send(msg)
											//}
										}
										msg.Text = "上车成功，请发送 /check 查看您的上车情况，发送 /checkcode 查看码子是否在助力池内！"
									} else {
										msg.Text = "车上乘客已满，请等待下一班次！"
									}
								} else {
									msg.Text = "您提交的总数超过了上限，不为贪得无厌的你添加进数据库！"
								}
							}
						} else {
							msg.Text = "请在按照格式输入\n\n格式：/city code1&code2&code3"
						}
					} else {
						msg.Text = "请在按照格式输入\n\n格式：/city code1&code2&code3"
					}
				case "bind":
					bindtexts := update.Message.Text
					bindtext := strings.Split(bindtexts, " ")
					fmt.Print(len(bindtext))
					if len(bindtext) != 2 {
						msg.Text = "正确的格式为 /bind 你的ptpin"
					} else if strings.Contains(bindtexts, "&") || strings.Contains(bindtexts, "'") || strings.Contains(bindtexts, ";") || strings.Contains(bindtexts, "=") || strings.Contains(bindtexts, "\"") {
						msg.Text = "存在非法字符！"
					} else {
						var bindedpin string
						err := db.QueryRow("SELECT binded FROM user WHERE tgid=?", update.Message.Chat.ID).Scan(&bindedpin)
						if err == sql.ErrNoRows {
							_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
							msg.Text = "您未注册！现在已经为您自动注册了！请重新执行！"
						} else {
							bindedpins := strings.Split(bindedpin, "&")
							isHad := 0
							for i := 0; i < len(bindedpins); i++ {
								if bindedpins[i] == bindtext[1] {
									isHad = 1
								}
							}
							if isHad == 1 {
								msg.Text = fmt.Sprintf("%v已被您绑定", bindtext[1])
							} else {
								if len(bindedpins)-1 >= 3 {
									msg.Text = "您的绑定已达到上限！"
								} else {
									_, _ = db.Exec("UPDATE user SET binded=? WHERE tgid=?", bindedpin+bindtext[1]+"&", update.Message.Chat.ID)
									_, _ = rs.Do("SADD", "pin", bindtext[1])
									msg.Text = "绑定成功！请输入 /bindlist 查看您绑定的账号"
								}
							}
						}
					}
				case "bindlist":
					var bindedpin string
					err := db.QueryRow("SELECT binded FROM user WHERE tgid=?", update.Message.Chat.ID).Scan(&bindedpin)
					if err == sql.ErrNoRows {
						_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
						msg.Text = "您未注册！现在已经为您自动注册了！请重新执行！"
					} else {
						bindedpins := strings.Split(bindedpin, "&")
						msg.Text = "您绑定的账号有：\n"
						for i := 0; i < len(bindedpins); i++ {
							msg.Text += bindedpins[i]
							msg.Text += "\n"
						}
					}
				case "unbind":
					bindtexts := update.Message.Text
					bindtext := strings.Split(bindtexts, " ")
					if len(bindtext) != 2 {
						msg.Text = "正确的格式为 /unbind 你的ptpin"
					} else if strings.Contains(bindtexts, "&") || strings.Contains(bindtexts, "@") || strings.Contains(bindtexts, "'") || strings.Contains(bindtexts, "\"") {
						msg.Text = "存在非法字符！"
					} else {
						var bindedpin string
						err := db.QueryRow("SELECT binded FROM user WHERE tgid=?", update.Message.Chat.ID).Scan(&bindedpin)
						if err == sql.ErrNoRows {
							_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
							msg.Text = "您未注册！现在已经为您自动注册了！请重新执行！"
						} else {
							bindedpins := strings.Split(bindedpin, "&")
							isHad := 0
							newbind := ""
							for i := 0; i < len(bindedpins); i++ {
								if bindedpins[i] == bindtext[1] {
									isHad = 1
								} else {
									if bindedpins[i] != "" {
										newbind = newbind + bindedpins[i] + "&"
									}
								}
							}
							if isHad == 0 {
								msg.Text = "您未绑定该ptpin！"
							} else {
								_, _ = db.Exec("UPDATE user SET binded=? WHERE tgid=?", newbind, update.Message.Chat.ID)
								_, _ = rs.Do("SREM", "pin", bindtext[1])
								msg.Text = "解绑成功！"
							}
						}
					}
				case "check":
					var bean, ddfactory, farm, health, jxfactory, pet, sgmh, carnivalcity int
					err := db.QueryRow("SELECT bean,ddfactory,farm,health,jxfactory,pet,sgmh,city FROM user where tgid=?", update.Message.Chat.ID).Scan(&bean, &ddfactory, &farm, &health, &jxfactory, &pet, &sgmh, &carnivalcity)
					if err == sql.ErrNoRows {
						_, _ = db.Exec("INSERT INTO user (tgid,isGroup) VALUES (?,0)", update.Message.Chat.ID)
						msg.Text = "您未注册！现在已经为您自动注册了！请重新上车！"
					} else {
						msg.Text = fmt.Sprintf("您的上车情况为：\n\n种豆得豆：%v\n东东工厂：%v\n东东农场：%v\n健康社区：%v\n京喜工厂：%v\n东东萌宠：%v\n闪购盲盒：%v\n城城：%v", bean, ddfactory, farm, health, jxfactory, pet, sgmh, carnivalcity)
					}
				case "checkcode":
					text := update.Message.Text
					textlist := strings.Split(text, " ")
					if len(textlist) == 3 {
						codename := textlist[1]
						if codename == "bean" || codename == "ddfactory" || codename == "farm" || codename == "health" || codename == "jxfactory" || codename == "pet" || codename == "sgmh" || codename == "cfd" || codename == "carnivalcity" || codename == "jxmc" || codename == "city" {
							result, _ := rs.Do("SISMEMBER", codename, textlist[2])
							resultrow := fmt.Sprintf("%v", result)
							//msg.Text=resultrow
							//bot.Send(msg)
							if resultrow == "1" {
								msg.Text = textlist[2] + "目前已在数据库中"
							} else {
								msg.Text = textlist[2] + "目前不在数据库中"
							}
						} else {
							msg.Text = "目前只支持bean, ddfactory, farm, health, jxfactory, pet,sgmh,cfd,carnivalcity,jxmc"
						}
					} else {
						msg.Text = "请按照格式搜索\n\n格式为：/checkcode (bean|ddfactory|farm|health|jxfactory|pet|sgmh|cfd|carnivalcity|jxmc) yourcode"
					}
				case "total":
					bean, _ := rs.Do("SCARD", "bean")
					ddfactory, _ := rs.Do("SCARD", "ddfactory")
					farm, _ := rs.Do("SCARD", "farm")
					health, _ := rs.Do("SCARD", "health")
					jxfactory, _ := rs.Do("SCARD", "jxfactory")
					pet, _ := rs.Do("SCARD", "pet")
					sgmh, _ := rs.Do("SCARD", "sgmh")
					cfd, _ := rs.Do("SCARD", "cfd")
					jxmc, _ := rs.Do("SCARD", "jxmc")
					pin, _ := rs.Do("SCARD", "pin")
					carnivalcity, _ := rs.Do("SCARD", "city")
					msg.Text = fmt.Sprintf("当前上车人数为：\n种豆得豆：%v\n东东工厂：%v\n东东农场：%v\n健康社区：%v\n京喜工厂：%v\n东东萌宠：%v\n闪购盲盒：%v\n财富岛（自动车）：%v\n惊喜牧场（自动车）：%v\n已绑定Telegram账号的京东账号总数：%v\n城城：%v", bean, ddfactory, farm, health, jxfactory, pet, sgmh, cfd, jxmc, pin, carnivalcity)
				case "redisclear":
					//fmt.Printf(strconv.FormatInt(update.Message.Chat.ID, 10))
					if update.Message.Chat.ID == 5021721171 {
						result, _ := rs.Do("FLUSHDB")
						msg.Text = fmt.Sprintf("%v", result)
						bot.Send(msg)
						_, _ = db.Exec("UPDATE user SET bean=0 , ddfactory=0 , farm=0 , health=0 , jxfactory=0 , pet=0,sgmh=0,cfd=0,carnivalcity=0")
					} else {
						msg.Text = "你没有权限！"
					}
				default:
					msg.Text = "命令不存在"
				}
				rs.Close()
				bot.Send(msg)
			} else {
				//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			}
		}
	}
}

func GetCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}
