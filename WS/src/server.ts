import { WebSocketServer, WebSocket } from "ws";
import type { ConfigType } from "./config/Config";
import Config from "./config/Config";
import type Redis from "ioredis";
import RedisClient from "./redis/Redis";
import { UserConnectionMap, MarketStreamWsConnectionMap } from "./connections/connectionMap"
import Types from "./types";
import MarketSubscriber from "./redis/Subscriber";

class Server {
  private conf: ConfigType;
  public Server: WebSocketServer
  private redisInstance: RedisClient | undefined
  private redisClient: Redis | undefined;
  private userConnMap: UserConnectionMap;
  private marketConnMap: MarketStreamWsConnectionMap;
  private marketSubscriber: MarketSubscriber

  constructor() {
    this.conf = new Config().MustLoad();
    this.redisInstance = RedisClient.getInstance();
    this.Server = new WebSocketServer({ port: this.conf.PORT });
    this.redisClient = RedisClient.getInstance().getClient();
    this.marketConnMap = new MarketStreamWsConnectionMap();
    this.marketSubscriber = new MarketSubscriber(this.redisClient, this.marketConnMap)
    this.userConnMap = new UserConnectionMap()
  }

  public InitServer() {
    this.Server.on("connection", this.connectionHandler.bind(this))
    console.info("WS Server started on port ", this.conf.PORT)
  }


  private connectionHandler(ws: WebSocket) {

    ws.on("message", (data) => {
      try {
        console.info("New Message recieved on WS server");
        const Message = JSON.parse(data.toString());
      } catch (error) {
        console.error("Invalid JSON data please provide a valid Data")
        ws.send(JSON.stringify(Types.ErrorMessage.InvalidJSON))

      }
    })
  }


  private messageHandler(message: any) { // add proper typing for message

  }
}




export default Server;
