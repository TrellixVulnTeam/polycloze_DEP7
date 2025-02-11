import { ActivityHistory } from "./schema";

import {
    CategoryScale,
    Chart,
    ChartData,
    ChartDataset,
    Colors,
    Filler,
    Legend,
    LineController,
    LineElement,
    LinearScale,
    PointElement,
    Tooltip,
} from "chart.js";

Chart.register(
    CategoryScale,
    Colors,
    Filler,
    Legend,
    LineController,
    LineElement,
    LinearScale,
    PointElement,
    Tooltip,
);

Chart.defaults.font.family = "Nunito";

const dayLabels = [
    "Sun",
    "Mon",
    "Tue",
    "Wed",
    "Thu",
    "Fri",
    "Sat",
];

// Returns date n days ago.
function dateNDaysAgo(n: number): Date {
    const daysSinceEpoch = Date.now() / 1000 / 60 / 60 / 24 - n;
    return new Date(daysSinceEpoch * 24 * 60 * 60 * 1000);
}

// Computes vocabulary size at each day this year.
export function computeVocabularySize({ activities, aggregates }: ActivityHistory): number[] {
    const vocab = new Array(367).fill(0);

    // Set increments
    for (let i = 0; i < activities.length; i++) {
        const { learned, forgotten } = activities[i];
        vocab[i] = learned - forgotten;
    }
    vocab[vocab.length - 1] = aggregates.learned - aggregates.forgotten;

    // Accumulate vocab size
    for (let i = vocab.length - 2; i >= 0; i--) {
        vocab[i] += vocab[i + 1];
    }
    return vocab;
}

// Returns vocabulary size data over the past week.
function vocabularyData(activityHistory: ActivityHistory): ChartData {
    const week = activityHistory.activities.slice(0, 7);
    const labels = week.map((_, i) => dayLabels[dateNDaysAgo(i).getDay()]).reverse();
    const data = computeVocabularySize(activityHistory).slice(0, labels.length).reverse();
    return {
        labels,
        datasets: [
            {
                data,
                label: "Vocabulary size",
                cubicInterpolationMode: "monotone",
                fill: true,
            },
        ],
    };
}

// Creates dataset for activity data.
function createDataset(label: string, data: number[]): ChartDataset {
    return {data, label, cubicInterpolationMode: "monotone"};
}

function activityData(activityHistory: ActivityHistory): ChartData {
    const week = activityHistory.activities.slice(0, 7);
    const labels = week.map((_, i) => dayLabels[dateNDaysAgo(i).getDay()]).reverse();
    return {
        labels,
        datasets: [
            createDataset("Learned", week.map(a => a.learned).reverse()),
            createDataset("Forgotten", week.map(a => a.forgotten).reverse()),
            createDataset("Unimproved", week.map(a => a.unimproved).reverse()),
            createDataset("Crammed", week.map(a => a.crammed).reverse()),
            createDataset("Strengthened", week.map(a => a.strengthened).reverse()),
        ],
    };
}

function createChart(canvas: HTMLCanvasElement, activityHistory: ActivityHistory): Chart {
    return new Chart(canvas, {
        type: "line",
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    min: 0,
                    ticks: {
                        stepSize: 1,
                    },
                },
            },
            plugins: {
                legend: {
                    position: "bottom",
                },
            },
        },
        data: vocabularyData(activityHistory),
    });
}

// Wraps around chart to make it responsive.
// See https://www.chartjs.org/docs/latest/configuration/responsive.html.
function createChartContainer(chart: HTMLCanvasElement): HTMLDivElement {
    const div = document.createElement("div");
    div.style.position = "relative";
    div.appendChild(chart);
    return div;
}

export function createVocabularyChart(activityHistory: ActivityHistory): HTMLDivElement {
    const canvas = document.createElement("canvas");
    createChart(canvas, activityHistory);
    return createChartContainer(canvas);
}

export function createActivityChart(activityHistory: ActivityHistory): HTMLDivElement {
    const canvas = document.createElement("canvas");
    new Chart(canvas, {
        type: "line",
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: "bottom",
                },
            },
            scales: {
                y: {
                    ticks: {
                        stepSize: 1,
                    },
                },
            },
        },
        data: activityData(activityHistory),
    });
    return createChartContainer(canvas);
}
