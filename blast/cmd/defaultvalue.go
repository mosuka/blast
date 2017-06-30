//  Copyright (c) 2017 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

const (
	DefaultConfig                string = ""
	DefaultLogFormat             string = "text"
	DefaultLogOutput             string = ""
	DefaultLogLevel              string = "info"
	DefaultPort                  int    = 10000
	DefaultPath                  string = "/var/indigo/data/index"
	DefaultIndexMapping          string = ""
	DefaultIndexType             string = "upside_down"
	DefaultKvstore               string = "boltdb"
	DefaultKvconfig              string = ""
	DefaultDeleteIndexAtStartup  bool   = false
	DefaultDeleteIndexAtShutdown bool   = false
)
