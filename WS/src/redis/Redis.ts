import Redis from "ioredis";

export default class RedisClient {
  private static instance: RedisClient;
  private client: Redis;

  private constructor() {
    this.client = new Redis(process.env.REDIS_URL as string);

    this.client.on("connect", () => {
      console.log("Redis connected");
    });

    this.client.on("error", (err) => {
      console.error("Redis connection error:", err);
    });
  }

  public static getInstance(): RedisClient {
    if (!RedisClient.instance) {
      RedisClient.instance = new RedisClient();
    }

    return RedisClient.instance;
  }

  public getClient(): Redis {
    return this.client;
  }
}
