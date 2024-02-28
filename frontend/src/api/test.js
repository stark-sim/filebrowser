import { fetchURL } from "./utils";

export function cdDownload(data) {
  return fetchURL("/api/cd/download", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
}
