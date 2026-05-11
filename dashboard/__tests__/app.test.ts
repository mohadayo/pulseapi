import request from "supertest";
import { createApp } from "../src/app";

const app = createApp();

describe("Dashboard API", () => {
  describe("GET /health", () => {
    it("returns healthy status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.service).toBe("dashboard");
      expect(res.body.status).toBe("healthy");
      expect(res.body.timestamp).toBeDefined();
    });
  });

  describe("GET /alerts", () => {
    it("returns alerts list", async () => {
      const res = await request(app).get("/alerts");
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });
  });

  describe("POST /alerts", () => {
    it("creates an alert with valid data", async () => {
      const res = await request(app)
        .post("/alerts")
        .send({ target: "web-app", message: "High latency detected", severity: "warning" });
      expect(res.status).toBe(201);
      expect(res.body.target).toBe("web-app");
      expect(res.body.message).toBe("High latency detected");
      expect(res.body.severity).toBe("warning");
      expect(res.body.acknowledged).toBe(false);
      expect(res.body.id).toBeDefined();
    });

    it("defaults severity to info", async () => {
      const res = await request(app)
        .post("/alerts")
        .send({ target: "api", message: "Routine check" });
      expect(res.status).toBe(201);
      expect(res.body.severity).toBe("info");
    });

    it("rejects alert without required fields", async () => {
      const res = await request(app)
        .post("/alerts")
        .send({ severity: "critical" });
      expect(res.status).toBe(400);
      expect(res.body.error).toBeDefined();
    });
  });

  describe("PATCH /alerts/:id/acknowledge", () => {
    it("acknowledges an existing alert", async () => {
      const createRes = await request(app)
        .post("/alerts")
        .send({ target: "db", message: "Connection pool low", severity: "critical" });
      const alertId = createRes.body.id;

      const res = await request(app).patch(`/alerts/${alertId}/acknowledge`);
      expect(res.status).toBe(200);
      expect(res.body.acknowledged).toBe(true);
    });

    it("returns 404 for nonexistent alert", async () => {
      const res = await request(app).patch("/alerts/nonexistent/acknowledge");
      expect(res.status).toBe(404);
    });
  });

  describe("DELETE /alerts/:id", () => {
    it("deletes an existing alert", async () => {
      const createRes = await request(app)
        .post("/alerts")
        .send({ target: "cache", message: "Cache miss rate high" });
      const alertId = createRes.body.id;

      const res = await request(app).delete(`/alerts/${alertId}`);
      expect(res.status).toBe(204);
    });

    it("returns 404 for nonexistent alert", async () => {
      const res = await request(app).delete("/alerts/nonexistent");
      expect(res.status).toBe(404);
    });
  });

  describe("GET /summary", () => {
    it("returns alert summary", async () => {
      const res = await request(app).get("/summary");
      expect(res.status).toBe(200);
      expect(res.body.total).toBeDefined();
      expect(res.body.acknowledged).toBeDefined();
      expect(res.body.unacknowledged).toBeDefined();
      expect(res.body.by_severity).toBeDefined();
    });
  });
});
