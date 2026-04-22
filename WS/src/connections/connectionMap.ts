import WebSocket from "ws";

class UserConnectionMap {

  public UserConnectionMap: Map<string, WebSocket>
  constructor() {
    this.UserConnectionMap = new Map<string, WebSocket>();
  }
  public AddUser(userId: string, ws: WebSocket) {
    this.UserConnectionMap.set(userId, ws);
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

    this.MarketConnectionMap.get(marketID)?.add(ws)
  }

}

export { UserConnectionMap, MarketStreamWsConnectionMap };
