import { Chart, ChartConfiguration } from "chart.js/auto";
import "chartjs-adapter-date-fns";
import { fromUnixTime, getUnixTime, subHours } from "date-fns";
import { enUS } from "date-fns/locale";

type Data = { time: number; count: number }[];

const chart_containers = Array.from(document.getElementsByClassName("chart"));

var charts: { [n: number]: Chart<"line", { x: Date; y: number }[]> } = {};
const urlParams = new URLSearchParams(window.location.search);

for (let chart_container of chart_containers) {
  const canvas = <HTMLCanvasElement>(
    Array.from(chart_container.children).find((e) => e.tagName === "CANVAS")
  );
  const data: Data = JSON.parse(atob(canvas.getAttribute("data")));
  const current_time = new Date();

  const config: ChartConfiguration<"line", { x: Date; y: number }[]> = {
    type: "line",
    options: {
      scales: {
        y: {
          beginAtZero: true,
        },
        x: {
          type: "time",
          suggestedMax: current_time.valueOf(),
          min: subHours(
            current_time,
            urlParams.has("history") ? Number(urlParams.get("history")) : 5
          ).valueOf(),
          time: {
            unit: "hour",
          },
          adapters: {
            date: {
              locale: enUS,
            },
          },
        },
      },
    },
    data: {
      datasets: [
        {
          label: "player count",
          data: data.map((entry) => {
            return { x: fromUnixTime(entry.time), y: entry.count };
          }),
        },
      ],
    },
  };
  charts[Number(canvas.getAttribute("server-id"))] = new Chart(canvas, config);
}

const ws = new WebSocket("/ws");

type ServerStatusUpdate = {
  server_id: number;
  time: number;
  count: number;
};

ws.addEventListener("message", (msg) => {
  const updates: ServerStatusUpdate[] = JSON.parse(msg.data);

  for (const update of updates) {
    const chart = charts[update.server_id];
    chart.data.datasets[0].data.push({
      x: fromUnixTime(update.time),
      y: update.count,
    });
    chart.update();
  }
});
