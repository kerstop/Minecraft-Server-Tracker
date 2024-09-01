import { Chart, ChartConfiguration } from "chart.js/auto";
import "chartjs-adapter-date-fns";
import { fromUnixTime, getUnixTime, subHours } from "date-fns";
import { enUS } from "date-fns/locale";

type Data = { time: number; count: number }[];

const charts = Array.from(
  document.getElementsByClassName("chart")
);

for (let chart of charts) {
  const canvas = <HTMLCanvasElement>Array.from(chart.children).find(e => e.tagName === "CANVAS")
  const data: Data = JSON.parse(atob(canvas.getAttribute("data")));
  const current_time = new Date();

  const config: ChartConfiguration<"line", number[], Date> = {
    type: "line",
    options: {
      scales: {
        y: {
          beginAtZero: true,
        },
        x: {
          type: "time",
          max: current_time.valueOf(),
          min: subHours(current_time, 5).valueOf(),
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
      xLabels: data.map((entry) => fromUnixTime(entry.time)),
      datasets: [
        {
          label: "player count",
          data: data.map((entry) => entry.count),
        },
      ],
    },
  };
  new Chart(canvas, config);
}
