import { createApp } from "./app";

const port = parseInt(process.env.DASHBOARD_PORT || "8082", 10);

const app = createApp();

app.listen(port, () => {
  console.log(`[INFO] Dashboard service started on port ${port}`);
});
