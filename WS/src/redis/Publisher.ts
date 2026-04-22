import type Redis from "ioredis";

class Publisher {

  private publisher: Redis;
  constructor(redisClient: Redis) {
    this.publisher = redisClient;

  }

  public async PublishMessageInMarket(marketID: string, message: string) {
    await this.publisher.publish(marketID, message);
    console.log("message published to marketID", marketID, "Message \n ", message);
  }
}


export default Publisher;
