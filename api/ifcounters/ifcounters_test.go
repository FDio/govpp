// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ifcounters

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestDecodeCounters(t *testing.T) {
	RegisterTestingT(t)

	testCounters := VnetInterfaceCounters{
		VnetCounterType: 1,
		IsCombined:      0,
		FirstSwIfIndex:  5,
		Count:           2,
		Data: []byte{0, 0, 0, 2, // Count
			0, 0, 0, 0, 0, 0, 0, 10, // first counter
			0, 0, 0, 0, 0, 0, 0, 11}, // second counter
	}
	counters, err := DecodeCounters(testCounters)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(len(counters)).To(BeEquivalentTo(2), "Incorrect size of the returned slice.")

	Expect(counters[0].Type).To(BeEquivalentTo(1), "Incorrect counter type.")
	Expect(counters[0].SwIfIndex).To(BeEquivalentTo(5), "Incorrect SwIfIndex.")
	Expect(counters[0].Packets).To(BeEquivalentTo(10), "Incorrect Packets count.")

	Expect(counters[1].Type).To(BeEquivalentTo(1), "Incorrect counter type.")
	Expect(counters[1].SwIfIndex).To(BeEquivalentTo(6), "Incorrect SwIfIndex.")
	Expect(counters[1].Packets).To(BeEquivalentTo(11), "Incorrect Packets count.")
}

func TestDecodeCombinedCounters(t *testing.T) {
	RegisterTestingT(t)

	testCounters := VnetInterfaceCounters{
		VnetCounterType: 1,
		IsCombined:      1,
		FirstSwIfIndex:  20,
		Count:           2,
		Data: []byte{0, 0, 0, 2, // Count
			0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 11, // first counter
			0, 0, 0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 13}, // second counter
	}
	counters, err := DecodeCombinedCounters(testCounters)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(len(counters)).To(BeEquivalentTo(2), "Incorrect size of the returned slice.")

	Expect(counters[0].Type).To(BeEquivalentTo(1), "Incorrect counter type.")
	Expect(counters[0].SwIfIndex).To(BeEquivalentTo(20), "Incorrect SwIfIndex.")
	Expect(counters[0].Packets).To(BeEquivalentTo(10), "Incorrect Packets count.")
	Expect(counters[0].Bytes).To(BeEquivalentTo(11), "Incorrect Bytes count.")

	Expect(counters[1].Type).To(BeEquivalentTo(1), "Incorrect counter type.")
	Expect(counters[1].SwIfIndex).To(BeEquivalentTo(21), "Incorrect SwIfIndex.")
	Expect(counters[1].Packets).To(BeEquivalentTo(12), "Incorrect Packets count.")
	Expect(counters[1].Bytes).To(BeEquivalentTo(13), "Incorrect Bytes count.")
}

func TestDecodeCountersNegative1(t *testing.T) {
	RegisterTestingT(t)

	testCounters := VnetInterfaceCounters{
		IsCombined: 1, // invalid, should be 0
	}
	counters, err := DecodeCounters(testCounters)

	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid argument"))
	Expect(counters).To(BeNil())
}

func TestDecodeCombinedCountersNegative1(t *testing.T) {
	RegisterTestingT(t)

	testCounters := VnetInterfaceCounters{
		IsCombined: 0, // invalid, should be 1
	}
	counters, err := DecodeCombinedCounters(testCounters)

	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid argument"))
	Expect(counters).To(BeNil())
}

func TestDecodeCountersNegative2(t *testing.T) {
	RegisterTestingT(t)

	testCounters := VnetInterfaceCounters{
		IsCombined: 0,
		// no data
	}
	counters, err := DecodeCounters(testCounters)

	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("unable to decode"))
	Expect(counters).To(BeNil())
}

func TestDecodeCombinedCountersNegative2(t *testing.T) {
	RegisterTestingT(t)

	testCounters := VnetInterfaceCounters{
		IsCombined: 1,
		// no data
	}
	counters, err := DecodeCombinedCounters(testCounters)

	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("unable to decode"))
	Expect(counters).To(BeNil())
}
