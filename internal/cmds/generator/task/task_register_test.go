/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/**
"ThisIsASampleCamelCaseString",
"IPAddress",
"ConvertIPAndHTML",
"SimpleCase",
"PDFLoader",
"ASimpleXMLParser",
*/

var _ = Describe("Test", func() {
	It("Test", func() {
		Expect(camelCaseToSpaceStyle("ChangeTenantRootPasswordFlow")).To(Equal("change tenant root password flow"))
		Expect(camelCaseToSpaceStyle("ModifyClusterSpec")).To(Equal("modify cluster spec"))
		Expect(camelCaseToSpaceStyle("RestartServers")).To(Equal("restart servers"))
		Expect(camelCaseToSpaceStyle("CamelCaseStringIPExample")).To(Equal("camel case string ipexample"))
		Expect(camelCaseToSpaceStyle("ThisIsASampleCamelCaseString")).To(Equal("this is asample camel case string"))
		Expect(camelCaseToSpaceStyle("IPAddress")).To(Equal("ipaddress"))
		Expect(camelCaseToSpaceStyle("ConvertIPAndHTML")).To(Equal("convert ipand html"))
		Expect(camelCaseToSpaceStyle("SimpleCase")).To(Equal("simple case"))
		Expect(camelCaseToSpaceStyle("PDFLoader")).To(Equal("pdfloader"))
		Expect(camelCaseToSpaceStyle("ASimpleXMLParser")).To(Equal("asimple xmlparser"))
	})
})
