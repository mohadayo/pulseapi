import pytest
from fastapi.testclient import TestClient

from monitor.app import app, targets


@pytest.fixture(autouse=True)
def clear_targets():
    targets.clear()
    yield
    targets.clear()


@pytest.fixture
def client():
    return TestClient(app)


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["service"] == "monitor"
    assert data["status"] == "healthy"
    assert "timestamp" in data


def test_list_targets_empty(client):
    resp = client.get("/targets")
    assert resp.status_code == 200
    assert resp.json() == []


def test_add_target(client):
    resp = client.post("/targets", json={
        "name": "test-svc",
        "url": "http://localhost:9999",
        "interval": 10,
    })
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == "test-svc"
    assert data["url"].rstrip("/") == "http://localhost:9999"
    assert data["interval"] == 10


def test_add_duplicate_target(client):
    client.post("/targets", json={
        "name": "dup",
        "url": "http://localhost:9999",
    })
    resp = client.post("/targets", json={
        "name": "dup",
        "url": "http://localhost:9998",
    })
    assert resp.status_code == 409


def test_remove_target(client):
    client.post("/targets", json={
        "name": "to-remove",
        "url": "http://localhost:9999",
    })
    resp = client.delete("/targets/to-remove")
    assert resp.status_code == 204

    resp = client.get("/targets")
    assert len(resp.json()) == 0


def test_remove_nonexistent_target(client):
    resp = client.delete("/targets/nonexistent")
    assert resp.status_code == 404


def test_check_target_unreachable(client):
    client.post("/targets", json={
        "name": "unreachable",
        "url": "http://localhost:1",
    })
    resp = client.post("/targets/unreachable/check")
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"] == "unreachable"
    assert data["status"] == "unreachable"
    assert "checked_at" in data


def test_check_nonexistent_target(client):
    resp = client.post("/targets/ghost/check")
    assert resp.status_code == 404


def test_list_targets_after_add(client):
    client.post("/targets", json={
        "name": "svc-a",
        "url": "http://localhost:9999",
    })
    client.post("/targets", json={
        "name": "svc-b",
        "url": "http://localhost:9998",
    })
    resp = client.get("/targets")
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 2
    names = {t["name"] for t in data}
    assert names == {"svc-a", "svc-b"}
