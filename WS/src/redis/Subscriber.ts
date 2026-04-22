import type Redis from "ioredis";
import type { MarketStreamWsConnectionMap } from "../connections/connectionMap";

class MarketSubscriber {

  private subscriber: Redis;
  private marketMapConn: MarketStreamWsConnectionMap;
  private subscribedMarkets: Set<string>
  constructor(redisClient: Redis, marketMap: MarketStreamWsConnectionMap) {
    this.subscriber = redisClient.duplicate();
    this.marketMapConn = marketMap;
    this.subscribedMarkets = new Set<string>();
    this.subscriber.on("message", this.handleMessage.bind(this))
  }

  public async SubscribeMarket(marketId: string, userId: string) {

    try {
      if (this.subscribedMarkets.has(marketId)) return;
      await this.subscriber.subscribe(marketId);
      this.subscribedMarkets.add(marketId);
    } catch (error) {
      console.error("ERROR in subscribe market of market ID :: ", marketId)
    }
  }

  private handleMessage(channel: string, message: string) {
    const clients = this.marketMapConn.MarketConnectionMap.get(channel);

    if (!clients) { return; }

    for (const client of clients) {
      if (client.readyState == client.OPEN) {
        client.send("Subscribed message recieved" + message);
      }
    }
  }

}


export default MarketSubscriber;
