import type Redis from "ioredis";

class Publisher {

  private publisher: Redis;
  constructor(redisClient: Redis) {
    this.publisher = redisClient;

  }

  public async PublishMessageInMarket(marketID: string, message: string) {
    try {

      await this.publisher.publish("MARKET_" + marketID, message);

      console.log("message published to marketID", marketID, "Message \n ", message);
    }
    catch (err) {

      console.error("Error in publishing the message in the market stream :: ", err)
    }
  }
}

export default Publisher;
