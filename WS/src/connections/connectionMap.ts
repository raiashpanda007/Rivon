import { randomUUIDv7 } from "bun";
import WebSocket from "ws";

class UserConnectionMap {

  public UserConnectionMap: Map<string, WebSocket>
  constructor() {
    this.UserConnectionMap = new Map<string, WebSocket>();
  }
  public AddUser(userId: string, ws: WebSocket) {
    this.UserConnectionMap.set(userId, ws);
  }
  public RemoveUser(userId: string) {
    this.UserConnectionMap.delete(userId);
  }

}

type ConnectionMeta = { userId: string; marketId: string };

class ConnectionMap {
  private ConnectionMap: Map<string, WebSocket>
  private meta: Map<string, ConnectionMeta>

  constructor() {
    this.ConnectionMap = new Map<string, WebSocket>();
    this.meta = new Map<string, ConnectionMeta>();
  }

  public AddConnection(ws: WebSocket): string {
    const id = randomUUIDv7();
    this.ConnectionMap.set(id, ws);
    return id;
  }

  public Get(connectionId: string): WebSocket | undefined {
    return this.ConnectionMap.get(connectionId);
  }

  public SetMeta(connectionId: string, userId: string, marketId: string) {
    this.meta.set(connectionId, { userId, marketId });
  }

  public GetMeta(connectionId: string): ConnectionMeta | undefined {
    return this.meta.get(connectionId);
  }

  public Remove(connectionId: string) {
    this.ConnectionMap.delete(connectionId);
    this.meta.delete(connectionId);
  }
}


class MarketStreamWsConnectionMap {
  public MarketConnectionMap: Map<string, Set<WebSocket>>
  constructor() {
    this.MarketConnectionMap = new Map<string, Set<WebSocket>>();
  }


  Add(marketID: string, ws: WebSocket) {
    if (!this.MarketConnectionMap.has(marketID)) {
      this.MarketConnectionMap.set(marketID, new Set<WebSocket>());
    }
    this.MarketConnectionMap.get(marketID)?.add(ws);
  }

  Remove(marketID: string, ws: WebSocket) {
    this.MarketConnectionMap.get(marketID)?.delete(ws);
  }

}

export { UserConnectionMap, MarketStreamWsConnectionMap, ConnectionMap };
