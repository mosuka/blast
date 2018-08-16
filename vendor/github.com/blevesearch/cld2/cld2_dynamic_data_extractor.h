// Copyright 2014 Google Inc. All Rights Reserved.                                                  
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

#ifndef CLD2_INTERNAL_CLD2_DYNAMIC_DATA_EXTRACTOR_H_
#define CLD2_INTERNAL_CLD2_DYNAMIC_DATA_EXTRACTOR_H_

#include "cld2_dynamic_data.h"
#include "integral_types.h"
#include "cld2tablesummary.h"
#include "utf8statetable.h"
#include "scoreonescriptspan.h"

namespace CLD2DynamicDataExtractor {

// Enable or disable debugging; 0 to disable, 1 to enable
void setDebug(int debug);

// Populates all the UTF8-related fields of the header, and returns the total
// space required within the binary blob to represent the non-primitive data.
void initUtf8Headers(CLD2DynamicData::FileHeader* header,
  const CLD2::UTF8PropObj* utf8Object);

// Populates all the AvgDeltaOctaScore-related fields of the header.
void initDeltaHeaders(CLD2DynamicData::FileHeader* header,
  const CLD2::uint32 deltaLength);

// Populates all fields of all table headers for the specified table summaries.
// Tables are laid out back-to-back in the order that they are specified in the
// input array of summaries, and the headers are filled in in the same order.
// IMPORTANT: The Supplement data structure must contain exactly one entry in
// indirectTableSizes for each CLD2TableSummary in the summaries parameter,
// in the same order.
void initTableHeaders(const CLD2::CLD2TableSummary** summaries,
  const int numSummaries, 
  const CLD2DynamicData::Supplement* supplement,
  CLD2DynamicData::TableHeader* tableSummaryHeaders);

// Align all entries in the data block along boundaries that are multiples of
// the specified number of bytes. For example, to align everything along 64-bit
// boundaries, pass an alignment of 8 (bytes).
void alignAll(CLD2DynamicData::FileHeader* header, const int alignment);

// Write the dynamic data file to disk.
void writeDataFile(const CLD2::ScoringTables* data,
  const CLD2DynamicData::Supplement* supplement,
  const char* fileName);

} // End namespace CLD2DynamicDataExtractor
#endif  // CLD2_INTERNAL_CLD2_DYNAMIC_DATA_EXTRACTOR_H_
