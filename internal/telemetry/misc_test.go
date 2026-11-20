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
	"context"
	"net/http"
	"net/http/httptest"
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

	DescribeTable("HEAD succeeds with HTTP and HTTPS", func(newServer func(http.Handler) *httptest.Server) {
		server := newServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		DeferCleanup(server.Close)
		resp, err := server.Client().Head(server.URL)
		Expect(err).ShouldNot(HaveOccurred())
		DeferCleanup(resp.Body.Close)
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
	},
		Entry("HTTP", httptest.NewServer),
		Entry("HTTPS", httptest.NewTLSServer),
	)

	It("rejects a HEAD URL without a scheme", func() {
		_, err := http.DefaultClient.Head("www.baidu.com")
		Expect(err).Should(HaveOccurred())
	})

	It("returns a timeout when the HEAD endpoint does not respond", func() {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		DeferCleanup(server.Close)
		client := &http.Client{Timeout: time.Second}
		_, err := client.Head(server.URL)
		Expect(err).To(MatchError(context.DeadlineExceeded))
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
		Skip("This case will panic with a high probability.")
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
		Skip("This case will panic definitely.")
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
