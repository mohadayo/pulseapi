import express, { Request, Response } from "express";

interface Alert {
  id: string;
  target: string;
  message: string;
  severity: "info" | "warning" | "critical";
  created_at: string;
  acknowledged: boolean;
}

const alerts: Map<string, Alert> = new Map();
let alertCounter = 0;

function generateId(): string {
  alertCounter += 1;
  return `alert-${alertCounter}`;
}

export function createApp(): express.Application {
  const app = express();
  app.use(express.json());

  app.get("/health", (_req: Request, res: Response) => {
    console.log("[INFO] Health check requested");
    res.json({
      service: "dashboard",
      status: "healthy",
      timestamp: new Date().toISOString(),
    });
  });

  app.get("/alerts", (_req: Request, res: Response) => {
    console.log(`[INFO] Listing ${alerts.size} alerts`);
    const list = Array.from(alerts.values());
    res.json(list);
  });

  app.post("/alerts", (req: Request, res: Response) => {
    const { target, message, severity } = req.body;

    if (!target || !message) {
      console.log("[WARN] Missing required fields in alert creation");
      res.status(400).json({ error: "target and message are required" });
      return;
    }

    const validSeverities = ["info", "warning", "critical"];
    const sev = severity && validSeverities.includes(severity) ? severity : "info";

    const alert: Alert = {
      id: generateId(),
      target,
      message,
      severity: sev,
      created_at: new Date().toISOString(),
      acknowledged: false,
    };

    alerts.set(alert.id, alert);
    console.log(`[INFO] Created alert ${alert.id} for target '${target}' [${sev}]`);
    res.status(201).json(alert);
  });

  app.patch("/alerts/:id/acknowledge", (req: Request, res: Response) => {
    const id = req.params.id as string;
    const alert = alerts.get(id);
    if (!alert) {
      console.log(`[WARN] Alert '${id}' not found`);
      res.status(404).json({ error: "alert not found" });
      return;
    }

    alert.acknowledged = true;
    console.log(`[INFO] Alert '${id}' acknowledged`);
    res.json(alert);
  });

  app.delete("/alerts/:id", (req: Request, res: Response) => {
    const id = req.params.id as string;
    if (!alerts.has(id)) {
      console.log(`[WARN] Alert '${id}' not found for deletion`);
      res.status(404).json({ error: "alert not found" });
      return;
    }

    alerts.delete(id);
    console.log(`[INFO] Alert '${id}' deleted`);
    res.status(204).send();
  });

  app.get("/summary", (_req: Request, res: Response) => {
    const list = Array.from(alerts.values());
    const summary = {
      total: list.length,
      acknowledged: list.filter((a) => a.acknowledged).length,
      unacknowledged: list.filter((a) => !a.acknowledged).length,
      by_severity: {
        info: list.filter((a) => a.severity === "info").length,
        warning: list.filter((a) => a.severity === "warning").length,
        critical: list.filter((a) => a.severity === "critical").length,
      },
    };
    console.log("[INFO] Summary requested");
    res.json(summary);
  });

  return app;
}
