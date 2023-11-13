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

package telemetry

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Telemetry", Label("misc"), func() {
	It("Test url parse 1", func() {
		u, err := url.Parse("https://www.baidu.com")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(u.Scheme).Should(Equal("https"))
		Expect(u.Host).Should(Equal("www.baidu.com"))
	})

	It("Test url parse 2", func() {
		u, err := url.Parse("http://www.baidu.com")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(u.Scheme).Should(Equal("http"))
		Expect(u.Host).Should(Equal("www.baidu.com"))
	})

	It("Test url parse 3", func() {
		u, err := url.Parse("www.baidu.com")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(u.Scheme).Should(Equal(""))
		Expect(u.Host).Should(Equal(""))
	})

	It("Test url parse and string()", func() {
		urlStr := "https://www.baidu.com"
		u, err := url.Parse(urlStr)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(u.String()).Should(Equal(urlStr))
	})

	It("Test head request", func() {
		_, err := http.DefaultClient.Head("https://www.baidu.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Test head request 2", func() {
		_, err := http.DefaultClient.Head("http://www.baidu.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Test head request 3", func() {
		_, err := http.DefaultClient.Head("www.baidu.com")
		Expect(err).Should(HaveOccurred())
	})

	It("Test head a not exist url with timeout 1s", func() {
		clt := http.Client{
			Timeout: time.Second,
		}
		_, err := clt.Head("https://www.baidx.com/abc")
		Expect(err).Should(HaveOccurred())
	})

	It("Test concurrent map writes with sync.Map, 1000 times", Label("panic"), func() {
		// This case will not panic.
		var m sync.Map
		for i := 0; i < 1000; i++ {
			go func() {
				defer GinkgoRecover()
				m.Store("a", "1")
			}()
			go func() {
				defer GinkgoRecover()
				m.Store("b", "2")
			}()
		}
	})

	It("Test concurrent map writes 100 times", Label("panic"), func() {
		// This case will panic with a high probability.
		m := make(map[string]string)
		for i := 0; i < 100; i++ {
			go func() {
				defer GinkgoRecover()
				m["a"] = "1"
			}()
			go func() {
				defer GinkgoRecover()
				m["b"] = "2"
			}()
		}
	})

	It("Test concurrent map writes 1000 times", Label("panic"), func() {
		// This case will definitely panic.
		m := make(map[string]string)
		for i := 0; i < 1000; i++ {
			go func() {
				defer GinkgoRecover()
				m["a"] = "1"
			}()
			go func() {
				defer GinkgoRecover()
				m["b"] = "2"
			}()
		}
	})
})
