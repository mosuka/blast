// Copyright 2015 ikawaha
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dic

// Trie is an interface representing retrieval ability.
type Trie interface {
	Search(input string) []int32
	PrefixSearch(input string) (length int, output []int32)
	CommonPrefixSearch(input string) (lens []int, outputs [][]int32)
	CommonPrefixSearchCallback(input string, callback func(id, l int))
}
