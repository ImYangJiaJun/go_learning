package main

import (
	"fmt"
	"time"
)

func main() {
	timeObj := time.Now()
	fmt.Println(timeObj) //2026-08-15 23:29:18.4519978 +0800 CST m=+0.000000001

	year := timeObj.Year()
	month := timeObj.Month()
	day := timeObj.Day()
	hour := timeObj.Hour()
	minute := timeObj.Minute()
	second := timeObj.Second()

	fmt.Printf("%02d-%02d-%02d %02d:%02d:%02d\n", year, month, day, hour, minute, second) //2026-08-15 23:30:38

	//使用自带函数进行格式化
	/*
		格式化格式
		2006	年
		01		月
		02		日
		03/15	时（12/24小时制）
		04		分
		05		秒
	*/
	timeObjF := timeObj.Format("2006/01/02 15:04:05")
	fmt.Println(timeObjF)

	//获取当前时间戳
	fmt.Println("---------------------------------------------------------")
	unixTime := timeObj.Unix() //秒（毫秒用UnixMilli()）
	fmt.Println("Unix:", unixTime)
	unixNaTime := timeObj.UnixNano() //纳秒
	fmt.Println("UnixNano:", unixNaTime)

	//时间戳转换日期字符串
	fmt.Println("---------------------------------------------------------")
	timeObj2 := time.Unix(unixTime, 0)
	fmt.Println("From Unix:", timeObj2)
	timeObj3 := time.Unix(0, unixNaTime)
	fmt.Println("From UnixNano:", timeObj3)

	fmt.Println("---------------------------------------------------------")
	//日期字符串转换时间戳
	strTime := "2026/08/15 23:47:19"
	tmp := "2006/01/02 15:04:05"                                        //转换模板
	timeObj4, success := time.ParseInLocation(tmp, strTime, time.Local) //ParseInLocation(模板,要转换的字符串,时区) 返回：时间对象，error（nil表示转换成功）
	if success == nil {
		fmt.Println("ParseInLocation Success:", timeObj4)
		fmt.Println("TimeUnix", timeObj4.Unix())
	}

	fmt.Println("---------------------------------------------------------")
	/*
		time包中定义的时间间隔类型的常量
		const{
			Nanosecond	 Duration	= 1
			Microsecond				= 1000 * Nanosecond
			Millisecond(毫秒)		= 1000 * Microsecond
			Second					= 1000 * Millisecond
			Minute					= 60 * Second
			Hour					= 60 * Minute
		}

	*/
	//时间操作函数
	fmt.Println("Time Now:", timeObj)
	timeObj = timeObj.Add(time.Hour)
	fmt.Println("Time After 1H:", timeObj)

	//定时器
	fmt.Println("---------------------------------------------------------")
	ticker := time.NewTicker(time.Second)
	runTimes := 5
	for t := range ticker.C {
		if runTimes <= 0 {
			fmt.Println("end")
			ticker.Stop() //终止定时器
			break
		}
		fmt.Println("Ticker At:", t)
		runTimes--
	}

	//休眠方法
	fmt.Println("---------------------------------------------------------")
	runTimes = 5
	for {
		time.Sleep(time.Second)
		if runTimes <= 0 {
			fmt.Println("end")
			break
		}
		fmt.Println("Ticker At:", time.Now())
		runTimes--
	}

}
