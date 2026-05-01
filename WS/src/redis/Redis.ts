import Redis from "ioredis";
import type { ConfigType } from "../config/Config";
import Config from "../config/Config";

export default class RedisClient {
  private static instance: RedisClient;
  private client: Redis;

  private conf: ConfigType

  private constructor() {
    this.conf = new Config().MustLoad()
    this.client = new Redis(this.conf.REDIS_URL)

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
