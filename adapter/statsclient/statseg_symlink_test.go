//  Copyright (c) 2026 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package statsclient

import (
	"unsafe"

	. "github.com/onsi/gomega"

	"testing"
)

// A v2 symlink directory entry packs (targetIndex, itemIndex) into its 8-byte union:
// the low 32 bits are the target directory index, the high 32 bits the item index.
func TestGetSymlinkIndexesV2(t *testing.T) {
	RegisterTestingT(t)

	const targetIndex, itemIndex = uint32(0x1234), uint32(0x5678)
	entry := statSegDirectoryEntryV2{
		directoryType: 6, // Symlink
		unionData:     uint64(targetIndex) | uint64(itemIndex)<<32,
	}

	ss := &statSegmentV2{}
	gotTarget, gotItem := ss.GetSymlinkIndexes(dirSegment(unsafe.Pointer(&entry)))

	Expect(gotTarget).To(Equal(targetIndex))
	Expect(gotItem).To(Equal(itemIndex))
}

// v1 has no symlinks; the accessor is a no-op returning zeroes.
func TestGetSymlinkIndexesV1(t *testing.T) {
	RegisterTestingT(t)

	ss := &statSegmentV1{}
	target, item := ss.GetSymlinkIndexes(nil)

	Expect(target).To(BeEquivalentTo(0))
	Expect(item).To(BeEquivalentTo(0))
}
