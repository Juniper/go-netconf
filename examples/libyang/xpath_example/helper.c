/**
 * @file helper.c
 * @author Mislav Novakovic <mislav.novakovic@sartura.hr>
 * @brief implementation of helper function for go program.
 *
 * @copyright
 * Copyright 2017 Deutsche Telekom AG.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <libyang/libyang.h>
#include "helper.h"
#include "_cgo_export.h"

void CErrorCallback(LY_LOG_LEVEL level, const char *msg, const char *path) {
	return GoErrorCallback(level, (char *) msg, (char *) path);
}
