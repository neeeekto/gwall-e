package http

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("URL Utils", func() {
	Describe("normalizeURL", func() {
		It("should handle empty baseURL and path", func() {
			Expect(normalizeURL("", "")).To(Equal("/"))
		})

		It("should handle baseURL with trailing slash", func() {
			Expect(normalizeURL("http://example.com/", "path")).To(Equal("http://example.com/path"))
		})

		It("should handle path with leading slash", func() {
			Expect(normalizeURL("http://example.com", "/path")).To(Equal("http://example.com/path"))
		})

		It("should handle both trailing and leading slashes", func() {
			Expect(normalizeURL("http://example.com/", "/path")).To(Equal("http://example.com/path"))
		})

		It("should handle multiple segments in path", func() {
			Expect(normalizeURL("http://example.com", "path/to/resource")).To(Equal("http://example.com/path/to/resource"))
		})

		It("should handle empty path", func() {
			Expect(normalizeURL("http://example.com", "")).To(Equal("http://example.com/"))
		})

		It("should handle empty baseURL", func() {
			Expect(normalizeURL("", "path")).To(Equal("/path"))
		})
	})
})