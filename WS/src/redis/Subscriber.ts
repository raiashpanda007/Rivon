import type Redis from "ioredis";
import type { MarketStreamWsConnectionMap, ConnectionMap } from "../connections/connectionMap";

const UNICAST_TYPES = new Set(["ORDERBOOK_DATA", "DEPTH_DATA"]);

class MarketSubscriber {

  private subscriber: Redis;
  private marketMapConn: MarketStreamWsConnectionMap;
  private connMap: ConnectionMap;
  private subscribedMarkets: Set<string>;

  constructor(redisClient: Redis, marketMap: MarketStreamWsConnectionMap, connMap: ConnectionMap) {
    this.subscriber = redisClient.duplicate();
    this.marketMapConn = marketMap;
    this.connMap = connMap;
    this.subscribedMarkets = new Set<string>();
    this.subscriber.on("message", this.handleMessage.bind(this));
  }

  public async SubscribeMarket(marketId: string) {
    try {
      if (this.subscribedMarkets.has(marketId)) return;
      await this.subscriber.subscribe("WS_OUT_" + marketId);
      this.subscribedMarkets.add(marketId);
    } catch (error) {
      console.error("ERROR in subscribe market of market ID :: ", marketId);
    }
  }

  private handleMessage(channel: string, message: string) {
    const marketId = channel.startsWith("WS_OUT_") ? channel.slice(7) : channel;

    let parsed: { type?: string; connectionId?: string } = {};
    try {
      parsed = JSON.parse(message);
    } catch {
      // unparseable — fall through to broadcast
    }

    if (parsed.type && UNICAST_TYPES.has(parsed.type) && parsed.connectionId) {
      const ws = this.connMap.Get(parsed.connectionId);
      if (ws && ws.readyState === ws.OPEN) {
        ws.send(message);
      }
      return;
    }

    const clients = this.marketMapConn.MarketConnectionMap.get(marketId);
    if (!clients) return;

    for (const client of clients) {
      if (client.readyState === client.OPEN) {
        client.send(message);
      }
    }
  }

}


export default MarketSubscriber;
