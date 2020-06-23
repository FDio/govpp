//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package binapigen

import (
	"strings"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

const (
	serviceEventPrefix   = "want_"
	serviceDumpSuffix    = "_dump"
	serviceDetailsSuffix = "_details"
	serviceReplySuffix   = "_reply"
)

func validateService(svc vppapi.Service) {
	for _, rpc := range svc.RPCs {
		validateRPC(rpc)
	}
}

func validateRPC(rpc vppapi.RPC) {
	if len(rpc.Events) > 0 {
		// EVENT service
		if !strings.HasPrefix(rpc.RequestMsg, serviceEventPrefix) {
			logrus.Warnf("unusual EVENTS service: %+v\n"+
				"- events service %q does not have %q prefix in request.",
				rpc, rpc.Name, serviceEventPrefix)
		}
	} else if rpc.Stream {
		// STREAM service
		if !strings.HasSuffix(rpc.RequestMsg, serviceDumpSuffix) {
			logrus.Warnf("unusual STREAM service: %+v\n"+
				"- stream service %q does not have %q suffix in request.",
				rpc, rpc.Name, serviceDumpSuffix)
		}
		if !strings.HasSuffix(rpc.ReplyMsg, serviceDetailsSuffix) && !strings.HasSuffix(rpc.StreamMsg, serviceDetailsSuffix) {
			logrus.Warnf("unusual STREAM service: %+v\n"+
				"- stream service %q does not have %q suffix in reply or stream msg.",
				rpc, rpc.Name, serviceDetailsSuffix)
		}
	} else if rpc.ReplyMsg != "" {
		// REQUEST service
		// some messages might have `null` reply (for example: memclnt)
		if !strings.HasSuffix(rpc.ReplyMsg, serviceReplySuffix) {
			logrus.Warnf("unusual REQUEST service: %+v\n"+
				"- service %q does not have %q suffix in reply.",
				rpc, rpc.Name, serviceReplySuffix)
		}
	}
}
