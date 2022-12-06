// JSON schemas used by server.

import { Item } from "./item";

export type ItemsSchema = {
    items: Item[];
};

export type ReviewSchema = {
    success: boolean;
    frequencyClass: number;    // describes student's level
};

export type Language = {
  code: string;
  name: string;
  bcp47: string;
};

export type LanguagesSchema = {
  languages: Language[];
};

export type Course = {
    l1: Language;
    l2: Language;
};

export type CoursesSchema = {
    courses: Course[];
};

export type Word = {
  word: string;
  learned: string;
  reviewed: string;
  due: string;
  strength: number;
};

// from /<l1>/<l2>/vocab
export type VocabularySchema = {
  words: Word[];
};

export type Activity = {
  forgotten: number;
  unimproved: number;
  crammed: number;
  learned: number;
  strengthened: number;
};

// from /<l1>/<l2>/activity
export type ActivityHistory = {
  activities: Activity[]; // up to one year of activities
  aggregates: Activity;   // for > 1 year old
};

// Not to be confused with sentence.Sentence.
export type RandomSentence = {
  id: number;
  tatoebaID?: number;
  text: string;
};

export type RandomSentencesSchema = {
  "sentences": RandomSentence[];
}

// Used in JSON schema only (indexedDB object store uses a different schema).
export type SyncReviewSchema = {
    word: string;       // key
    learned: Date | string;      // default now
    reviewed: Date | string;     // default now
    interval: number;   // default 0 or 24 hours
    sequenceNumber: number;
};

export type SyncRequestSchema = {
  latest: number; // Largest sequence number of ACK'ed review.
  reviews: SyncReviewSchema[];
  difficultyStats: string;  // stringified stats table
  intervalStats: string;    // stringified stats table
};

// Schema of response from /api/sync/<l1>/<l2>
export type SyncResponseSchema = {
  reviews?: SyncReviewSchema[];
  difficultyStats?: string;
  intervalStats?: string;
};
