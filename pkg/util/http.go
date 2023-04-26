/*
Copyright (c) 2023 OceanBase
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

func HTTPGET(reqURL string, reqDatas ...map[string]interface{}) (int, map[string]interface{}) {
	req, _ := http.NewRequest("GET", reqURL, nil)
	if reqDatas != nil {
		q := req.URL.Query()
		for _, reqData := range reqDatas {
			for k, v := range reqData {
				q.Add(k, v.(string))
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	c := &http.Client{
		Timeout: 1 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		log.Println("perr: ", perr)
		return 0, nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if berr != nil {
		log.Println("berr: ", berr)
		return 0, nil
	}
	responseData := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responseData)
	if jerr != nil {
		log.Println("jerr: ", jerr)
		return res.StatusCode, nil
	}
	return res.StatusCode, responseData
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
	responseData := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responseData)
	if jerr != nil {
		log.Println(jerr)
		return res.StatusCode, nil
	}
	return res.StatusCode, responseData
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
	responseData := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responseData)
	if jerr != nil {
		log.Println(jerr)
		return nil
	}
	return responseData
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
	responseData := make(map[string]interface{})
	jerr := json.Unmarshal(resBody, &responseData)
	if jerr != nil {
		log.Println(jerr)
		return nil
	}
	return responseData
}
