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

#include "debug.h"
#include <stdio.h>
#include <string>

#include "cldutil.h"
#include "getonescriptspan.h"
#include "lang_script.h"

using namespace std;

namespace CLD2 {

// Debug output string of one unigram
string GetUniAt(const char* text) {
  string retval;
  retval.clear();
  int uni_len = UniLen(text);
  retval.append(text, uni_len);
  return retval;
}

// Debug output string of one bigram
string GetBiAt(const char* text) {
  string retval;
  retval.clear();
  int bi_len = BiLen(text);
  retval.append(text, bi_len);
  return retval;
}

// Debug output string of one quadgram, including underscores
string GetQuadAt(const char* text) {
  string retval;
  retval.clear();
  if (text[-1] == ' ') {retval.append("_");}
  int quad_len = QuadLen(text);
  retval.append(text, quad_len);
  if (text[quad_len] == ' ') {retval.append("_");}
  return retval;
}

// Debug output string of one octagram, including underscores
string GetOctaAt(const char* text) {
  string retval;
  retval.clear();
  if (text[-1] == ' ') {retval.append("_");}
  int octa_len = OctaLen(text);
  retval.append(text, octa_len);
  if (text[octa_len] == ' ') {retval.append("_");}
  return retval;
}

// Debug output string of two octagrams, including underscores
string GetOcta2At(const char* text) {
  string retval;
  retval.clear();
  if (text[-1] == ' ') {retval.append("_");}
  int octa_len = OctaLen(text);
  retval.append(text, octa_len);
  if (text[octa_len] == ' ') {retval.append("_");}
  text += (octa_len + 1);
  int octa2_len = OctaLen(text);
  retval.append(text, octa2_len);
  if (text[octa2_len] == ' ') {retval.append("_");}
  return retval;
}

// Debug output string of one formatted pslang,qprob pair
string FmtLP(ULScript ulscript, uint8 pslang, uint8 qprob) {
  string retval;
  retval.clear();
  Language lang = FromPerScriptNumber(ulscript, pslang);
  char temp[16];
  sprintf(temp, "%s.%d", LanguageCode(lang), qprob);
  retval.append(temp);
  return retval;
}

// Debug output string of one formatted langprob
// Returns "en.24&nbsp;fr.10&nbsp;es.4"
string GetLangProbTxt(const ScoringContext* scoringcontext, uint32 langprob) {
  /*const uint16* pslangtolang = scoringcontext->pslangtolang;*/
  string retval;
  retval.clear();
  uint8 prob123 = (langprob >> 0) & 0xff;
  const uint8* prob123_entry = LgProb2TblEntry(prob123);
  uint8 top1 = (langprob >> 8) & 0xff;
  if (top1 > 0) {
    retval.append(FmtLP(scoringcontext->ulscript,
                        top1, LgProb3(prob123_entry, 0)));
  }
  uint8 top2 = (langprob >> 16) & 0xff;
  if (top2 > 0) {
    if (!retval.empty()) {retval.append("~");}
    retval.append(FmtLP(scoringcontext->ulscript,
                        top2, LgProb3(prob123_entry, 1)));
  }
  uint8 top3 = (langprob >> 24) & 0xff;
  if (top3 > 0) {
    if (!retval.empty()) {retval.append("~");}
    retval.append(FmtLP(scoringcontext->ulscript,
                        top3, LgProb3(prob123_entry, 2)));
  }
  return retval;
}


// Debug output string of one or two formatted quadgram langprobs
string GetScoreTxt(const ScoringContext* scoringcontext,
                   const CLD2TableSummary* base_obj, int indirect) {
  string retval;
  retval.clear();
  if (indirect < static_cast<int>(base_obj->kCLDTableSizeOne)) {
    // Up to three languages at indirect
    uint32 langprob = base_obj->kCLDTableInd[indirect];
    retval.append(GetLangProbTxt(scoringcontext, langprob));
  } else {
    // Up to six languages at start + 2 * (indirect - start)
    indirect += (indirect - base_obj->kCLDTableSizeOne);
    uint32 langprob = base_obj->kCLDTableInd[indirect];
    uint32 langprob2 = base_obj->kCLDTableInd[indirect + 1];
    retval.append(GetLangProbTxt(scoringcontext, langprob));
    if (!retval.empty()) {retval.append("~");}
    retval.append(GetLangProbTxt(scoringcontext, langprob2));
  }
  return retval;
}


// 16 background colors, perhaps from the low 4 bits of the language number
static const int kLangBackground[16] = {
  0xffd8d8, 0xf8ffd8, 0xd8ffe7, 0xd8f3ff,
  0xefd8ff, 0xffd8eb, 0xfff7d8, 0xe3ffd8,
  0xd8ffff, 0xe3d8ff, 0xffd8f7, 0xffebd8,
  0xefffd8, 0xd8fff3, 0xd8e7ff, 0xf8d8ff,
};

// 16 text colors, perhaps from the high 4 bits of the language number
// 00..7f
static const int kLangColor[16] = {
  0x000000, 0x7f2f00, 0x7f5f00, 0x6f7f00,        // first 16 lang: black text
  0x3f7f00, 0x0f7f00, 0x007f1f, 0x007f4f,
  0x007f7f, 0x004f7f, 0x001f7f, 0x0f007f,
  0x3f007f, 0x6f007f, 0x7f005f, 0x7f002f,
};

static const int kUnscoredText = 0xb0b0b0;        // medium-light gray
static const int kUnscoredBackground = 0xffffff;  // white
static const int kIgnoremeText = 0x8090a0;        // medium-light green-gray
static const int kIgnoremeBackground = 0xffeecc;  // light orange
static const int kEnglishBackground = 0xfffff4;   // very light yellow

static int GetBackColor(Language lang, bool lighten) {
  int retval;
  if (lang == ENGLISH) {
    retval = kEnglishBackground;
  } else if (lang == UNKNOWN_LANGUAGE) {
    retval = kUnscoredBackground;
  } else if (lang == TG_UNKNOWN_LANGUAGE) {
    retval = kIgnoremeBackground;
  } else if (lang < 0) {
    retval = kUnscoredBackground;
  } else {
    retval = kLangBackground[lang & 0x0f];
  }
  if (lighten) {
    // Make 1/2 as far away from white
    retval = (retval >> 1) | 0x808080;
  }
  return retval;
}

static int GetTextColor(Language lang, bool lighten) {
  int retval;
  if (lang == UNKNOWN_LANGUAGE) {
    retval = kUnscoredText;
  } else if (lang == TG_UNKNOWN_LANGUAGE) {
    retval = kIgnoremeText;
  } else if (lang < 0) {
    retval = kUnscoredText;
  } else {
    retval = kLangColor[(lang >> 4) & 0x0f];
  }
  if (lighten) {
    // Make 1/2 as far away from white
    retval = (retval >> 1) | 0x808080;
  }
  return retval;
}

string GetPlainEscapedText(const string& txt) {
  string retval;
  retval.clear();
  for (int i = 0; i < static_cast<int>(txt.size()); ++i) {
    char c = txt[i];
    if (c == '\n') {
      retval.append(" ");
    } else if (c == '\r') {
      retval.append(" ");
    } else {
      retval.append(1, c);
    }
  }
  return retval;
}

string GetHtmlEscapedText(const string& txt) {
  string retval;
  retval.clear();
  for (int i = 0; i < static_cast<int>(txt.size()); ++i) {
    char c = txt[i];
    if (c == '<') {
      retval.append("&lt;");
    } else if (c == '>') {
      retval.append("&gt;");
    } else if (c == '&') {
      retval.append("&amp;");
    } else if (c == '\'') {
      retval.append("&apos;");
    } else if (c == '"') {
      retval.append("&quot;");
    } else if (c == '\n') {
      retval.append(" ");
    } else if (c == '\r') {
      retval.append(" ");
    } else {
      retval.append(1, c);
    }
  }
  return retval;
}

string GetColorHtmlEscapedText(Language lang, const string& txt) {
  char temp[64];
  sprintf(temp, " <span style=\"background:#%06X;color:#%06X;\">\n",
          GetBackColor(lang, false),
          GetTextColor(lang, false));
  string esc_txt = string(temp);
  esc_txt.append(GetHtmlEscapedText(txt));
  esc_txt.append("</span>");
  return esc_txt;
}

string GetLangColorHtmlEscapedText(Language lang, const string& txt) {
  char temp[64];
  sprintf(temp, "[%s]", LanguageCode(lang));
  string esc_txt = string(temp);
  esc_txt.append(GetColorHtmlEscapedText(lang, txt));
  return esc_txt;
}


// For showing one chunk
// Print debug output for one scored chunk
// Optionally print out per-chunk scoring information
// In degenerate cases, hitbuffer and cspan can be NULL
void CLD2_Debug(const char* text,
                int lo_offset,
                int hi_offset,
                bool more_to_come, bool score_cjk,
                const ScoringHitBuffer* hitbuffer,
                const ScoringContext* scoringcontext,
                const ChunkSpan* cspan,
                const ChunkSummary* chunksummary) {
  FILE* df = scoringcontext->debug_file;
  if (df == NULL) {return;}

  if (scoringcontext->flags_cld2_verbose &&
      (hitbuffer != NULL) &&
      (cspan != NULL) && (hitbuffer->next_linear > 0)) {
    int base_limit = cspan->chunk_base + cspan->base_len;
    for (int i = cspan->chunk_base; i < base_limit; ++i) {
      int ngram_start = hitbuffer->linear[i].offset;
      uint32 langprob = hitbuffer->linear[i].langprob;
      string ngram_text;
      switch (hitbuffer->linear[i].type) {
      case UNIHIT:
        ngram_text = GetUniAt(&text[ngram_start]);
        break;
      case QUADHIT:
        ngram_text = GetQuadAt(&text[ngram_start]);
        break;
      case DELTAHIT:
      case DISTINCTHIT:
        if (score_cjk) {
          ngram_text = GetBiAt(&text[ngram_start]);
        } else {
          // TODO: figure out how to display optional two words
          ngram_text = GetOctaAt(&text[ngram_start]);
        }
        break;
      }
      string score_text = GetLangProbTxt(scoringcontext, langprob);
      fprintf(df, "%c:%s=%s&nbsp;&nbsp; ",
              "UQLD"[hitbuffer->linear[i].type],
              ngram_text.c_str(),
              score_text.c_str());
    }
    fprintf(df, "<br>\n");

    // Score boosts for langprior and distinct tokens
    // Get boosts for current script
    const LangBoosts* langprior_boost = &scoringcontext->langprior_boost.latn;
    const LangBoosts* langprior_whack = &scoringcontext->langprior_whack.latn;
    const LangBoosts* distinct_boost = &scoringcontext->distinct_boost.latn;
    if (scoringcontext->ulscript != ULScript_Latin) {
      langprior_boost = &scoringcontext->langprior_boost.othr;
      langprior_whack = &scoringcontext->langprior_whack.othr;
      distinct_boost = &scoringcontext->distinct_boost.othr;
    }
    fprintf(df, "LangPrior_boost: ");
    for (int k = 0; k < kMaxBoosts; ++k) {
      uint32 langprob = langprior_boost->langprob[k];
      if (langprob > 0) {
        fprintf(df, "%s&nbsp;&nbsp; ",
                GetLangProbTxt(scoringcontext, langprob).c_str());
      }
    }
    fprintf(df, "LangPrior_whack: ");
    for (int k = 0; k < kMaxBoosts; ++k) {
      uint32 langprob = langprior_whack->langprob[k];
      if (langprob > 0) {
        fprintf(df, "%s&nbsp;&nbsp; ",
                GetLangProbTxt(scoringcontext, langprob).c_str());
      }
    }
    fprintf(df, "Distinct_boost: ");
    for (int k = 0; k < kMaxBoosts; ++k) {
      uint32 langprob = distinct_boost->langprob[k];
      if (langprob > 0) {
        fprintf(df, "%s&nbsp;&nbsp; ",
                GetLangProbTxt(scoringcontext, langprob).c_str());
      }
    }
    fprintf(df, "<br>\n");

    // Print chunksummary
    fprintf(df, "%s.%d %s.%d %dB %d# %s %dRd %dRs<br>\n",
            LanguageCode(static_cast<Language>(chunksummary->lang1)),
            chunksummary->score1,
            LanguageCode(static_cast<Language>(chunksummary->lang2)),
            chunksummary->score2,
            chunksummary->bytes,
            chunksummary->grams,
            ULScriptCode(static_cast<ULScript>(chunksummary->ulscript)),
            chunksummary->reliability_delta,
            chunksummary->reliability_score);
  }   // End flags_cld2_verbose linear


  // Print annotated colored text of this chunk
  bool is_reliable = true;
  bool match_prior = false;
  int reliable = CLD2::minint(chunksummary->reliability_delta,
                             chunksummary->reliability_score);
  is_reliable = (reliable >= 75);
  match_prior = (chunksummary->lang1 == scoringcontext->prior_chunk_lang);
  if (!is_reliable) {match_prior = false;}

  if (match_prior) {
    fprintf(df, "[]");
  } else if (is_reliable) {
    fprintf(df, "[%s]",
            LanguageCode(static_cast<Language>(chunksummary->lang1)));
  } else {
    fprintf(df, "[%s*.%d/%s.%d]",
            LanguageCode(static_cast<Language>(chunksummary->lang1)),
            chunksummary->score1,
            LanguageCode(static_cast<Language>(chunksummary->lang2)),
            chunksummary->score2);
  }

  int chunktext_len = hi_offset - lo_offset;
  if (chunktext_len < 0) {
    chunktext_len = 0;
    fprintf(df, " LEN_ERR hi %d lo %d<br>\n", hi_offset, lo_offset);
  }
  string chunk_text(&text[lo_offset], chunktext_len);

  Language lang = static_cast<Language>(chunksummary->lang1);
  fprintf(df, " <span style=\"background:#%06X;color:#%06X;\">\n",
          GetBackColor(lang, false),
          GetTextColor(lang, false));
  fprintf(df, "%s", chunk_text.c_str());
  if (scoringcontext->flags_cld2_cr) {
    fprintf(df, "</span><br>\n");
  } else {
    fprintf(df, "</span> \n");
  }
}

// For showing all chunks
void CLD2_Debug2(const char* text,
                 bool more_to_come, bool score_cjk,
                 const ScoringHitBuffer* hitbuffer,
                 const ScoringContext* scoringcontext,
                 const SummaryBuffer* summarybuffer) {
  FILE* df = scoringcontext->debug_file;
  if (df == NULL) {return;}
  uint16 prior_chunk_lang = static_cast<uint16>(UNKNOWN_LANGUAGE);

  for (int i = 0; i < summarybuffer->n; ++i) {
    fprintf(df, "Debug2[%d] ", i);
    const ChunkSummary* chunksummary = &summarybuffer->chunksummary[i];
     // Print annotated colored text of this chunk
    bool is_reliable = true;
    bool match_prior = false;
    int reliable = CLD2::minint(chunksummary->reliability_delta,
                                chunksummary->reliability_score);
    is_reliable = (reliable >= 75);
    match_prior = (chunksummary->lang1 == prior_chunk_lang);
    if (!is_reliable) {match_prior = false;}

    if (match_prior) {
      fprintf(df, "[]");
    } else if (is_reliable) {
      fprintf(df, "[%s]",
              LanguageCode(static_cast<Language>(chunksummary->lang1)));
    } else {
      fprintf(df, "[%s*.%d/%s.%d]",
              LanguageCode(static_cast<Language>(chunksummary->lang1)),
              chunksummary->score1,
              LanguageCode(static_cast<Language>(chunksummary->lang2)),
              chunksummary->score2);
    }

    int lo_offset = chunksummary->offset;
    int chunktext_len = chunksummary->bytes;
    string chunk_text(&text[lo_offset], chunktext_len);

    Language lang = static_cast<Language>(chunksummary->lang1);
    fprintf(df, " <span style=\"background:#%06X;color:#%06X;\">\n",
            GetBackColor(lang, false),
            GetTextColor(lang, false));
    fprintf(df, "%s", chunk_text.c_str());
    if (scoringcontext->flags_cld2_cr) {
      fprintf(df, "</span><br>\n");
    } else {
      fprintf(df, "</span> \n");
    }
    prior_chunk_lang = chunksummary->lang1;
  }
}

void DumpResultChunkVector(FILE* f, const char* src,
                           ResultChunkVector* resultchunkvector) {
  fprintf(f, "DumpResultChunkVector[%ld]<br>\n", resultchunkvector->size());
  for (int i = 0; i < static_cast<int>(resultchunkvector->size()); ++i) {
    ResultChunk* rc = &(*resultchunkvector)[i];
    Language lang1 = static_cast<Language>(rc->lang1);
    string this_chunk = string(src, rc->offset, rc->bytes);
    fprintf(f, "[%d]{%d %d %s} ", i, rc->offset, rc->bytes, LanguageCode(lang1));
    fprintf(f, "%s<br>\n", GetColorHtmlEscapedText(lang1, this_chunk).c_str());
  }
  fprintf(f, "<br>\n");
}

}       // End namespace CLD2


