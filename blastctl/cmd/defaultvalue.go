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
	DefaultOutputFormat        string = "json"
	DefaultVersionFlag         bool   = false
	DefaultResource            string = ""
	DefaultServer              string = "localhost:10000"
	DefaultBatchSize           int32  = 1000
	DefaultIncludeIndexMapping bool   = false
	DefaultIncludeIndexType    bool   = false
	DefaultIncludeKvstore      bool   = false
	DefaultIncludeKvconfig     bool   = false
	DefaultId                  string = ""
	DefaultDocFields           string = ""
	DefaultQuery               string = ""
	DefaultSize                int    = 10
	DefaultFrom                int    = 0
	DefaultExplain             bool   = false
	DefaultFacets              string = ""
	DefaultHighlight           string = ""
	DefaultHighlightStyle      string = "html"
	DefaultIncludeLocations    bool   = false
)

var (
	DefaultFields          []string = []string{}
	DefaultSorts           []string = []string{}
	DefaultHighlightFields []string = []string{}
)
