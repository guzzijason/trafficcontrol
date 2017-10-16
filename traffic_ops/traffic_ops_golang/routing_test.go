package main

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"net/http"
	"testing"

	"fmt"

	"bytes"
	"context"
	"net/http/httptest"
)

func TestCreateRouteMap(t *testing.T) {
	authBase := AuthBase{false, "secret", nil, func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "authWasCalled", "true")
			handlerFunc(w, r.WithContext(ctx))
		}
	}}

	PathOneHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authWasCalled := getAuthWasCalled(ctx)

		fmt.Fprintf(w, "%s %s", "path1", authWasCalled)
	}

	PathTwoHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authWasCalled := getAuthWasCalled(ctx)

		fmt.Fprintf(w, "%s %s", "path2", authWasCalled)
	}

	routes := []Route{{1.2, http.MethodGet, `path1`, PathOneHandler, ServersPrivLevel, true, nil}, {1.2, http.MethodGet, `path2`, PathTwoHandler, 0, false, nil}}

	routeMap := CreateRouteMap(routes, authBase)

	route1Handler := routeMap["GET"][0].Handler

	w := httptest.NewRecorder()
	r, err := http.NewRequest("", "/", nil)
	if err != nil {
		t.Error("Error creating new request")
	}

	route1Handler(w, r)

	if bytes.Compare(w.Body.Bytes(), []byte("path1 true")) != 0 {
		t.Errorf("Got: %s \nExpected to receive path1 true\n", w.Body.Bytes())
	}

	route2Handler := routeMap["GET"][1].Handler

	w = httptest.NewRecorder()

	route2Handler(w, r)

	if bytes.Compare(w.Body.Bytes(), []byte("path2 false")) != 0 {
		t.Errorf("Got: %s \nExpected to receive path2 false\n", w.Body.Bytes())
	}
}

func getAuthWasCalled(ctx context.Context) string {
	val := ctx.Value("authWasCalled")
	if val != nil {
		return val.(string)
	}
	return "false"
}