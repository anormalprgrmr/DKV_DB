package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database", Ordered, func() {
	Context("Testing http endpoints", func() {
		masterPort := 8081
		replicaPorts := []int{8082, 8083}

		_, err := PutObject(masterPort, "foo", "bar")
		Expect(err).ToNot(HaveOccurred())

		for _, p := range replicaPorts {
			resp, err := GetObject(p, "foo")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Body).To(MatchJSON(`{"value":"bar"}`))
		}
	})
})
