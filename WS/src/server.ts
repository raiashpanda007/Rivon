import { WebSocketServer, WebSocket } from "ws";
import type { ConfigType } from "./config/Config";
import Config from "./config/Config";
import type Redis from "ioredis";
import RedisClient from "./redis/Redis";
import { UserConnectionMap, MarketStreamWsConnectionMap, ConnectionMap } from "./connections/connectionMap"
import Types, { ParseClient, PUBSLISHED_MESSAGE_TYPES } from "./types";
import MarketSubscriber from "./redis/Subscriber";
import Publisher from "./redis/Publisher";
import MessageHandler from "./handlers/MessageHandler";

class Server {
  private conf: ConfigType;
  public Server: WebSocketServer
  private redisClient: Redis | undefined;
  private userConnMap: UserConnectionMap;
  private marketConnMap: MarketStreamWsConnectionMap;
  private connMap: ConnectionMap;
  private marketSubscriber: MarketSubscriber
  private marketPublisher: Publisher
  private messageHandler: MessageHandler;
  constructor() {
    this.conf = new Config().MustLoad();
    this.Server = new WebSocketServer({ port: this.conf.PORT });
    this.redisClient = RedisClient.getInstance().getClient();
    this.marketConnMap = new MarketStreamWsConnectionMap();
    this.connMap = new ConnectionMap();
    this.marketSubscriber = new MarketSubscriber(this.redisClient, this.marketConnMap, this.connMap)
    this.userConnMap = new UserConnectionMap()
    this.marketPublisher = new Publisher(this.redisClient);
    this.messageHandler = new MessageHandler(this.marketPublisher, this.marketSubscriber, this.userConnMap, this.marketConnMap, this.connMap)
  }

  public InitServer() {
    this.Server.on("connection", this.connectionHandler.bind(this))
    console.info("WS Server started on port ", this.conf.PORT)
  }


  private connectionHandler(ws: WebSocket) {
    const connectionId = this.connMap.AddConnection(ws);

    ws.on("message", (data) => {
      const ClientMessage = ParseClient(data)
      if (!ClientMessage) {
        ws.send(JSON.stringify(Types.ErrorMessage.InvalidJSON))
        return;
      }
      this.messageHandler.Handler(ClientMessage, ws, connectionId)
    })

    ws.on("close", async () => {
      const meta = this.connMap.GetMeta(connectionId);
      if (meta) {
        this.marketConnMap.Remove(meta.marketId, ws);
        if (meta.userId) {
          this.userConnMap.RemoveUser(meta.userId);
          await this.marketPublisher.PublishMessageInMarket(meta.marketId, JSON.stringify({
            MessageType: PUBSLISHED_MESSAGE_TYPES.walletEvict,
            UserId: meta.userId,
            ConnectionId: connectionId,
          }));
        }
      }
      this.connMap.Remove(connectionId);
    })
  }


}


export default Server;
