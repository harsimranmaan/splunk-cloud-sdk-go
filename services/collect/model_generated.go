/*
 * Copyright © 2019 Splunk, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"): you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 *
 * Collect Service
 *
 * With the Splunk Cloud Collect service, you can manage how data collection jobs ingest event and metric data.
 *
 * API version: v1beta1.5
 * Generated by: OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.
 */

package collect

// number of jobs deleted.
type DeleteJobsResponse struct {
	Count int32 `json:"count"`
}

type EventExtraField struct {
	// Field name
	Name string `json:"name"`
	// Field value
	Value string `json:"value"`
}

type Job struct {
	// The ID of the connector used in the job.
	ConnectorID string `json:"connectorID"`
	Name        string `json:"name"`
	// The configuration of the connector used in the job.
	Parameters  map[string]interface{} `json:"parameters"`
	ScalePolicy map[string]interface{} `json:"scalePolicy"`
	// The cron schedule, in UTC time format.
	Schedule         string  `json:"schedule"`
	CreateUserID     *string `json:"createUserID,omitempty"`
	CreatedAt        *string `json:"createdAt,omitempty"`
	Id               *string `json:"id,omitempty"`
	LastModifiedAt   *string `json:"lastModifiedAt,omitempty"`
	LastUpdateUserID *string `json:"lastUpdateUserID,omitempty"`
}

type JobPatch struct {
	// The ID of the connector used in the job.
	ConnectorID *string `json:"connectorID,omitempty"`
	// The job name
	Name *string `json:"name,omitempty"`
	// The configuration of the connector used in the job.
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	ScalePolicy map[string]interface{} `json:"scalePolicy,omitempty"`
	// The cron schedule, in UTC time format.
	Schedule *string `json:"schedule,omitempty"`
}

type JobsPatch struct {
	// The ID of the connector used in the job.
	ConnectorID      *string           `json:"connectorID,omitempty"`
	EventExtraFields []EventExtraField `json:"eventExtraFields,omitempty"`
	ScalePolicy      *ScalePolicy      `json:"scalePolicy,omitempty"`
}

// List of jobs.
type ListJobsResponse struct {
	Data []Job `json:"data,omitempty"`
}

// The metadata for the patch jobs operation.
type Metadata struct {
	// The number of jobs that failed to update.
	Failures int64 `json:"failures"`
	// The number of jobs which match the query criteria.
	TotalMatchJobs int64 `json:"totalMatchJobs"`
}

type ModelError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// The optional details of the error.
	Details map[string]interface{} `json:"details,omitempty"`
	// An optional link to a web page with more information on the error.
	MoreInfo *string `json:"moreInfo,omitempty"`
}

type PatchJobResult struct {
	// The Job ID.
	Id string `json:"id"`
	// Successfully updated or not.
	Updated bool        `json:"updated"`
	Error   *ModelError `json:"error,omitempty"`
}

type PatchJobsResponse struct {
	Data     []PatchJobResult `json:"data"`
	Metadata Metadata         `json:"metadata"`
}

type ScalePolicy struct {
	Static StaticScale `json:"static"`
}

type SingleJobResponse struct {
	Data Job `json:"data"`
}

type StaticScale struct {
	// The number of collect workers.
	Workers int32 `json:"workers"`
}