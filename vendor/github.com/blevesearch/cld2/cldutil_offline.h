// Copyright 2013 Google Inc. All Rights Reserved.
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

//
// Author: dsites@google.com (Dick Sites)
//
// Just the stuff shared between offline table builder and online detector
//

#ifndef I18N_ENCODINGS_CLD2_INTERNAL_NEW_CLDUTIL_OFFLINE_H__
#define I18N_ENCODINGS_CLD2_INTERNAL_NEW_CLDUTIL_OFFLINE_H__

#include "cldutil_shared.h"
#include "integral_types.h"
#include "lang_script.h"

namespace CLD2 {

//------------------------------------------------------------------------------
// Offline: used by mapreduce or table construction
//------------------------------------------------------------------------------

// Find top two langs and scores for one word; underpins delta tables
void DoWordScore(const char* isrc, int srclen, ULScript ulscript,
                 const CLD2TableSummary* wrt_unigram_obj,
                 const CLD2TableSummary* wrt_quadgram_obj,
                 Language* lang1, int* score1,
                 Language* lang2, int* score2);

// For constructing tables
// Given a vector of 3 probabilities 1..12, find subscript of best table match.
// Minimizes RMS error
// Brute-force version
uint8 FindBestProb3Match(const uint8* prob3);

// Not sure who calls this...
// Return the probability for given language, or 0
int GetProb(Language lang, uint32 probs);


// Converts a unigram prob/lang byte into an approximate prob/lang triple
// Just keeps the largest value.
// ONLY used in mapreduce, doing refinement of language boundaries.
uint32 ApproxProb3(int propval);


// Take three packed languages and three probabilities 1..12 and put into uint32
// For offline construction of tables
uint32 ProbPackV2(uint8* plang3, uint8* prob3);

// Take uint32 and unpack into three packed languages and three probabilities
// For runtime use of tables
void ProbUnpackV2(uint32 prob, uint8* plang3, uint8* prob3);

}       // End namespace CLD2

#endif  // I18N_ENCODINGS_CLD2_INTERNAL_NEW_CLDUTIL_OFFLINE_H__



