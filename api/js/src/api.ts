// Wrappers for api calls.

import { Item } from "./item";
import { getL1, getL2 } from "./language";
import { fetchJson, resolve, submitJson } from "./request";
import {
    ActivityHistory,
    Course,
    CoursesSchema,
    ItemsSchema,
    Language,
    LanguagesSchema,
    RandomSentence,
    RandomSentencesSchema,
    ReviewSchema,
    Word,
    VocabularySchema,
} from "./schema";

type FetchVocabularyOptions = {
    // Path params
    l1?: string;    // L1 code
    l2?: string;    // L2 code

    // Search params
    limit?: number;  // Max number of items to fetch
    after?: string;  // Last item to exclude from query
    sortBy?: "word" | "reviewed" | "due" | "strength";
};

function defaultFetchVocabularyOptions(): FetchVocabularyOptions {
    return {
        l1: getL1().code,
        l2: getL2().code,
        limit: 50,
        after: "",
        sortBy: "word",
    };
}

export async function fetchVocabulary(options: FetchVocabularyOptions = {}): Promise<Word[]> {
    const { l1, l2, limit, after, sortBy } = {...defaultFetchVocabularyOptions(), ...options};
    const url = resolve(`/${l1}/${l2}/vocab`);
    setParams(url, { after, limit, sortBy });

    const json = await fetchJson<VocabularySchema>(url, {
        mode: "cors" as RequestMode,
    });
    return json.words || [];
}

type FetchActivityHistoryOptions = {
    // Path params
    l1?: string;
    l2?: string;
};

function defaultFetchActivityHistoryOptions(): FetchActivityHistoryOptions {
    return {
        l1: getL1().code,
        l2: getL2().code,
    };
}

// Fetches student's activity history over the past year.
export async function fetchActivityHistory(options: FetchActivityHistoryOptions = {}): Promise<ActivityHistory> {
    const {l1, l2} = {...defaultFetchActivityHistoryOptions(), ...options};
    const url = resolve(`/${l1}/${l2}/activity`);
    return await fetchJson<ActivityHistory>(url, {
        mode: "cors" as RequestMode,
    });
}

export async function fetchCourses(): Promise<Course[]> {
    const url = resolve("/api/courses");
    setParams(url, { t: "20221114" });
    const json = await fetchJson<CoursesSchema>(url, {
        mode: "cors" as RequestMode,
    });
    return json.courses;
}

// Fetches list of supported languages (L1).
export async function fetchLanguages(): Promise<Language[]> {
    const url = resolve("/api/languages");
    setParams(url, { t: "20221114" });
    const json = await fetchJson<LanguagesSchema>(url, {
        mode: "cors" as RequestMode,
    });
    return json.languages;
}

type FetchItemsOptions = {
    // Path params
    l1?: string;    // L1 code
    l2?: string;    // L2 code

    // Search params
    n?: number;      // Max number of items to fetch
    x?: string[];    // Words to exclude
};

function defaultFetchItemsOptions(): FetchItemsOptions {
    return {
        l1: getL1().code,
        l2: getL2().code,
        n: 10,
        x: [],
    };
}

export async function fetchItems(options: FetchItemsOptions = {}): Promise<Item[]> {
    const { l1, l2, n, x } = {...defaultFetchItemsOptions(), ...options};
    const url = resolve(`/${l1}/${l2}`);
    setParams(url, { n, x });

    const json = await fetchJson<ItemsSchema>(url, {
        mode: "cors" as RequestMode,
    });
    return json.items;
}

type FetchSentencesOptions = {
    l1?: string;
    l2?: string;
    limit?: number;
};

function defaultFetchSentencesOptions(): FetchSentencesOptions {
    return {
        l1: getL1().code,
        l2: getL2().code,
        limit: 1,
    };
}

export async function fetchSentences(options: FetchSentencesOptions = {}): Promise<RandomSentence[]> {
    const { l1, l2, limit } = {...defaultFetchSentencesOptions(), ...options};
    if (l1 == null || l2 == null) {
        throw new Error("l1 and l2 required");
    }

    const url = resolve("/api/sentences");
    setParams(url, {l1, l2, limit});

    const json = await fetchJson<RandomSentencesSchema>(url, {
        mode: "cors" as RequestMode,
    });
    return json.sentences;
}

type Params = {
    [name: string]: unknown;
};

function setParams(url: URL, params: Params) {
    for (const name of Object.getOwnPropertyNames(params)) {
        const value = params[name];
        if (value === undefined) {
            continue;
        }
        if (value instanceof Array) {
            for (const item of value) {
                url.searchParams.append(name, item);
            }
            continue;
        }
        url.searchParams.set(name, String(value));
    }
}

export function submitReview(word: string, correct: boolean): Promise<ReviewSchema> {
    const l1 = getL1().code;
    const l2 = getL2().code;

    const url = resolve(`/${l1}/${l2}`);
    const data = {
        reviews: [
            { word, correct },
        ],
    };
    return submitJson<ReviewSchema>(url, data);
}
