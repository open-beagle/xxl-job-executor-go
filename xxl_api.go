package xxl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var client *http.Client

func init() {
	client = &http.Client{}
}

type xxlApi struct {
	Options
	log    logger
	cookie string
}

type xxlExecutor struct {
	Id           int      `json:"id,omitempty"`
	Appname      string   `json:"appname,omitempty"`
	Title        string   `json:"title,omitempty"`
	AddressType  string   `json:"addressType,omitempty"`
	AddressList  string   `json:"addressList,omitempty"`
	UpdateTime   string   `json:"updateTime,omitempty"`
	RegistryList []string `json:"registryList,omitempty"`
}

type xxlJob struct {
	JobGroup        int    `json:"jobGroup,omitempty"`
	ScheduleConf    string `json:"scheduleConf,omitempty"`
	ExecutorHandler string `json:"executorHandler,omitempty"`
	JobDesc         string `json:"jobDesc,omitempty"`
	Id              int    `json:"id,omitempty"`
	Author          string `json:"author,omitempty"`
}

type webResExecutor struct {
	RecordsFiltered int           `json:"recordsFiltered,omitempty"`
	Data            []xxlExecutor `json:"data,omitempty"`
	RecordsTotal    int           `json:"recordsTotal,omitempty"`
}

type webResJob struct {
	RecordsFiltered int      `json:"recordsFiltered,omitempty"`
	Data            []xxlJob `json:"data,omitempty"`
	RecordsTotal    int      `json:"recordsTotal,omitempty"`
}

type wenResCode struct {
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Content string `json:"content,omitempty"`
}

func newXxlApi(opt Options) *xxlApi {
	xxl := &xxlApi{Options: opt}
	return xxl
}

func (x *xxlApi) login() error {
	sendURL := fmt.Sprintf("%s/login", x.Options.ServerAddr)
	header := make(map[string]string)
	header["Content-Type"] = "application/x-www-form-urlencoded"
	body := url.Values{}
	body.Add("userName", "admin")
	body.Add("password", x.Options.AdminPwd)
	resp, err := http.Post(sendURL, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
	if err != nil {
		return err
	} else {
		cookie := resp.Header.Get("Set-Cookie")
		x.cookie = cookie
		return nil
	}
}

// 检查并添加执行器
func (x *xxlApi) checkOrAddExecutor(appname, alias, addressList string) {
	executor, err := x.getExecutor(appname)
	if err != nil {
		x.log.Error("获取执行器错误:%s", err.Error())
	} else if executor.Appname == "" {
		x.addExecutor(appname, alias, addressList)
	} else if executor.Appname != "" && (executor.AddressList != addressList || executor.Title != alias) {
		x.updateExecutor(appname, alias, addressList, executor.Id)
	}
}

// 获取执行器
func (x *xxlApi) getExecutor(appname string) (executor xxlExecutor, err error) {
	// https://apaas5.wodcloud.com/xxl-job-admin/jobgroup/pageList
	if x.cookie == "" {
		if err := x.login(); err != nil {
			return executor, err
		}
	}
	sendURL := fmt.Sprintf("%s/jobgroup/pageList", x.Options.ServerAddr)
	body := url.Values{}
	body.Add("appname", appname)
	request, _ := http.NewRequest("POST", sendURL, strings.NewReader(body.Encode()))
	request.Header.Set("cookie", x.cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		return executor, err
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return executor, err
		}
		res := webResExecutor{}
		json.Unmarshal(respBody, &res)
		for _, v := range res.Data {
			if v.Appname == appname {
				return v, nil
			}
		}
		return executor, err
	}
}

// 添加执行器
func (x *xxlApi) addExecutor(appname, alias, addressList string) {
	// https://apaas5.wodcloud.com/xxl-job-admin/jobgroup/save
	if x.cookie == "" {
		if err := x.login(); err != nil {
			return
		}
	}
	sendURL := fmt.Sprintf("%s/jobgroup/save", x.Options.ServerAddr)
	body := url.Values{}
	body.Add("appname", appname)
	body.Add("title", alias)
	body.Add("addressType", "1")
	body.Add("addressList", addressList)
	request, _ := http.NewRequest("POST", sendURL, strings.NewReader(body.Encode()))
	request.Header.Set("cookie", x.cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		x.log.Error("调用接口【新增执行器】错误：%s", err.Error())
		return
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		// {"code":200,"msg":null,"content":null}
		if err != nil {
			x.log.Error("调用接口【新增执行器】返回错误：%s", err.Error())
			return
		}
		res := wenResCode{}
		json.Unmarshal(respBody, &res)
		if res.Code == 200 {
			return
		}
		x.log.Error("调用接口【新增执行器】错误信息：%s", res.Msg)
		return
	}
}

// 更新执行器
func (x *xxlApi) updateExecutor(appname, alias, addressList string, id int) {
	// https://apaas5.wodcloud.com/xxl-job-admin/jobgroup/save
	if x.cookie == "" {
		if err := x.login(); err != nil {
			return
		}
	}
	sendURL := fmt.Sprintf("%s/jobgroup/update", x.Options.ServerAddr)
	body := url.Values{}
	body.Add("appname", appname)
	body.Add("title", alias)
	body.Add("addressType", "1")
	body.Add("addressList", addressList)
	body.Add("id", strconv.Itoa(id))
	request, _ := http.NewRequest("POST", sendURL, strings.NewReader(body.Encode()))
	request.Header.Set("cookie", x.cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		x.log.Error("调用接口【新增执行器】错误：%s", err.Error())
		return
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		// {"code":200,"msg":null,"content":null}
		if err != nil {
			x.log.Error("调用接口【新增执行器】返回错误：%s", err.Error())
			return
		}
		res := wenResCode{}
		json.Unmarshal(respBody, &res)
		if res.Code == 200 {
			return
		}
		x.log.Error("调用接口【新增执行器】错误信息：%s", res.Msg)
		return
	}
}

// 检查并添加任务
func (x *xxlApi) checkOrAddJob(jobDesc, scheduleConf, executorHandler string) {
	job, err := x.getJob(executorHandler)
	if err != nil {
		x.log.Error("获取执行器错误:%s", err.Error())
	} else if job.ExecutorHandler == "" {
		x.addJob(jobDesc, scheduleConf, executorHandler)
		job, err = x.getJob(executorHandler)
		if err != nil {
			x.log.Error("获取执行器错误:%s", err.Error())
		}
	} else if job.ExecutorHandler == executorHandler && (job.JobDesc != jobDesc || job.ScheduleConf != scheduleConf) { //modify it if it is not equal.
		x.updateJob(jobDesc, scheduleConf, executorHandler, job.Id)
	}
	x.startJob(job.Id, executorHandler) //start job
}

// 获取任务
func (x *xxlApi) getJob(executorHandler string) (job xxlJob, err error) {
	// https://apaas5.wodcloud.com/xxl-job-admin/jobinfo/pageList
	if x.cookie == "" {
		if err := x.login(); err != nil {
			return job, err
		}
	}
	executor, err := x.getExecutor(x.RegistryKey)
	if err != nil {
		x.log.Error("获取执行器Id信息错误:%s", err.Error())
		return
	} else if executor.Id == 0 {
		x.log.Error("获取执行器Id为0")
		return
	}
	sendURL := fmt.Sprintf("%s/jobinfo/pageList", x.Options.ServerAddr)
	body := url.Values{}
	body.Add("jobGroup", strconv.Itoa(executor.Id))
	body.Add("executorHandler", executorHandler)
	body.Add("triggerStatus", "-1")
	request, _ := http.NewRequest("POST", sendURL, strings.NewReader(body.Encode()))
	request.Header.Set("cookie", x.cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		return job, err
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return job, err
		}
		res := webResJob{}
		json.Unmarshal(respBody, &res)
		for i, v := range res.Data {
			v.ExecutorHandler = strings.ReplaceAll(v.ExecutorHandler, " ", "")
			if v.JobGroup == executor.Id && v.ExecutorHandler == executorHandler {
				job = res.Data[i]
				return job, nil
			}
		}
		return job, err
	}
}

// 新增任务
func (x *xxlApi) addJob(jobDesc, scheduleConf, executorHandler string) {
	// https://apaas5.wodcloud.com/xxl-job-admin/jobinfo/pageList
	if x.cookie == "" {
		if err := x.login(); err != nil {
			return
		}
	}
	executor, err := x.getExecutor(x.RegistryKey)
	if err != nil {
		x.log.Error("获取执行器Id信息错误:%s", err.Error())
		return
	} else if executor.Id == 0 {
		x.log.Error("获取执行器Id为0")
		return
	}
	sendURL := fmt.Sprintf("%s/jobinfo/add", x.Options.ServerAddr)
	body := url.Values{}
	body.Add("jobGroup", strconv.Itoa(executor.Id))
	body.Add("jobDesc", jobDesc)
	body.Add("author", "beagle")
	body.Add("scheduleType", "CRON")
	body.Add("scheduleConf", scheduleConf)
	body.Add("cronGen_display", scheduleConf)
	body.Add("glueType", "BEAN")
	body.Add("executorHandler", executorHandler)
	body.Add("executorRouteStrategy", "FIRST")
	body.Add("misfireStrategy", "DO_NOTHING")
	body.Add("executorBlockStrategy", "SERIAL_EXECUTION")
	body.Add("executorTimeout", "0")
	body.Add("executorFailRetryCount", "0")
	body.Add("glueRemark", "GLUE代码初始化")
	request, _ := http.NewRequest("POST", sendURL, strings.NewReader(body.Encode()))
	request.Header.Set("cookie", x.cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		x.log.Error("调用接口【新增任务】错误：%s", err.Error())
		return
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			x.log.Error("调用接口【新增任务】返回错误：%s", err.Error())
			return
		}
		res := wenResCode{}
		json.Unmarshal(respBody, &res)
		if res.Code == 200 {
			return
		}
		x.log.Error("调用接口【新增任务】错误信息：%s", res.Msg)
		return
	}
}

// 修改任务
func (x *xxlApi) updateJob(jobDesc, scheduleConf, executorHandler string, id int) {
	// https://apaas5.wodcloud.com/xxl-job-admin/jobinfo/pageList
	if x.cookie == "" {
		if err := x.login(); err != nil {
			return
		}
	}
	executor, err := x.getExecutor(x.RegistryKey)
	if err != nil {
		x.log.Error("获取执行器Id信息错误:%s", err.Error())
		return
	} else if executor.Id == 0 {
		x.log.Error("获取执行器Id为0")
		return
	}
	sendURL := fmt.Sprintf("%s/jobinfo/update", x.Options.ServerAddr)
	body := url.Values{}
	body.Add("scheduleConf", scheduleConf)
	body.Add("jobDesc", jobDesc)
	body.Add("executorHandler", executorHandler)
	body.Add("id", strconv.Itoa(id))
	body.Add("author", "beagle")
	body.Add("scheduleType", "CRON")
	body.Add("executorRouteStrategy", "FIRST")
	body.Add("executorFailRetryCount", "0")
	body.Add("misfireStrategy", "DO_NOTHING")
	body.Add("executorBlockStrategy", "SERIAL_EXECUTION")
	body.Add("jobGroup", strconv.Itoa(executor.Id))
	body.Add("cronGen_display", scheduleConf)
	body.Add("glueType", "BEAN")
	body.Add("executorTimeout", "0")
	body.Add("triggerStatus", "1")
	body.Add("glueRemark", "GLUE代码初始化")
	request, _ := http.NewRequest("POST", sendURL, strings.NewReader(body.Encode()))
	request.Header.Set("cookie", x.cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		x.log.Error("调用接口【修改任务】错误：%s", err.Error())
		return
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			x.log.Error("调用接口【修改任务】返回错误：%s", err.Error())
			return
		}
		res := wenResCode{}
		json.Unmarshal(respBody, &res)
		if res.Code == 200 {
			return
		}
		x.log.Error("调用接口【修改任务】错误信息：%s", res.Msg)
		return
	}
}

// 启动任务
func (x *xxlApi) startJob(id int, executorHandler string) {
	if x.cookie == "" {
		if err := x.login(); err != nil {
			return
		}
	}
	executor, err := x.getExecutor(x.RegistryKey)
	if err != nil {
		x.log.Error("获取执行器Id信息错误:%s", err.Error())
		return
	} else if executor.Id == 0 {
		x.log.Error("获取执行器Id为0")
		return
	}
	sendURL := fmt.Sprintf("%s/jobinfo/start", x.Options.ServerAddr)
	body := url.Values{}
	body.Add("id", strconv.Itoa(id))
	request, _ := http.NewRequest("POST", sendURL, strings.NewReader(body.Encode()))
	request.Header.Set("cookie", x.cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		x.log.Error("调用接口【启动任务】错误：%s", err.Error())
		return
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			x.log.Error("调用接口【启动任务】返回错误：%s", err.Error())
			return
		}
		res := wenResCode{}
		json.Unmarshal(respBody, &res)
		if res.Code == 200 {
			return
		}
		x.log.Error("调用接口【启动任务】错误信息：%s", res.Msg)
		return
	}
}
