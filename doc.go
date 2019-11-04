//  Copyright (c) 2019 Cisco and/or its affiliates.
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

// Package govpp provides the entry point to govpp functionality. It provides the API for connecting the govpp core
// to VPP either using the default VPP adapter, or using the adapter previously set by SetAdapter function
// (useful mostly just for unit/integration tests with mocked VPP adapter).
//
// To create a connection to VPP, use govpp.Connect function:
//
//	conn, err := govpp.Connect()
//	if err != nil {
//		// handle error!
//	}
//	defer conn.Disconnect()
//
// Make sure you close the connection after using it. If the connection is not closed, it will leak resources. Please
// note that only one VPP connection is allowed for a single process.
//
// In case you need to mock the connection to VPP (e.g. for testing), use the govpp.SetAdapter function before
// calling govpp.Connect.
//
// Once connected to VPP, use the functions from the api package to communicate with it.
package govpp
