/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func HTTPGET(reqURL string) (int, map[string]interface{}) {
	req, _ := http.NewRequest("GET", reqURL, nil)
	c := &http.Client{
		Timeout: 1 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		log.Println(perr)
		return 0, nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if berr != nil {
		log.Println(berr)
		return 0, nil
	}
	responeDate := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responeDate)
	if jerr != nil {
		log.Println(jerr)
		return res.StatusCode, nil
	}
	return res.StatusCode, responeDate
}

func HTTPPOST(reqURL, reqData string) (int, map[string]interface{}) {
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(reqData))
	req.Header.Add("Content-Type", "application/json")
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		log.Println(perr)
		return 0, nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if berr != nil {
		log.Println(berr)
		return 0, nil
	}
	responeDate := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responeDate)
	if jerr != nil {
		log.Println(jerr)
		return res.StatusCode, nil
	}
	return res.StatusCode, responeDate
}

func HTTPPUT(reqURL, reqData string) map[string]interface{} {
	req, _ := http.NewRequest("PUT", reqURL, strings.NewReader(reqData))
	req.Header.Add("Content-Type", "application/json")
	c := &http.Client{
		Timeout: 1 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		log.Println(perr)
		return nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if berr != nil {
		log.Println(berr)
		return nil
	}
	responeDate := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responeDate)
	if jerr != nil {
		log.Println(jerr)
		return nil
	}
	return responeDate
}

func HTTPDELETE(reqURL, reqData string) map[string]interface{} {
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	c := &http.Client{
		Timeout: 1 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		log.Println(perr)
		return nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if berr != nil {
		log.Println(berr)
		return nil
	}
	responeDate := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responeDate)
	if jerr != nil {
		log.Println(jerr)
		return nil
	}
	return responeDate
}
