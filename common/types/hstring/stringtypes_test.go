// Copyright 2024 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hstring

import (
	"html/template"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/cast"
)

func TestRenderedString(t *testing.T) {
	c := qt.New(t)

	// Validate that it will behave like a string in Hugo settings.
	c.Assert(cast.ToString(HTML("Hugo")), qt.Equals, "Hugo")
	c.Assert(template.HTML(HTML("Hugo")), qt.Equals, template.HTML("Hugo"))
}
