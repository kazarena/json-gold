// Copyright 2015 Stanislav Nazarenko
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ld

// JsonLdApi is the main interface to JSON-LD API.
// See http://www.w3.org/TR/json-ld-api/ for detailed description of this interface.
type JsonLdApi struct {
}

// NewJsonLdApi creates a new instance of JsonLdApi and initialises it
// with the given JsonLdOptions structure.
func NewJsonLdApi() *JsonLdApi {
	return &JsonLdApi{}
}
