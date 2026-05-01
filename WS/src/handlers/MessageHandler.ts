import WebSocket from "ws";
import type { UserConnectionMap, MarketStreamWsConnectionMap, ConnectionMap } from "../connections/connectionMap";
import type Publisher from "../redis/Publisher";
import type MarketSubscriber from "../redis/Subscriber";
import { MessageType, type ClientMessage, PUBSLISHED_MESSAGE_TYPES } from "../types";

class MessageHandler {

  private publisher: Publisher;
  private marketSubscriber: MarketSubscriber;
  private userConnectionMap: UserConnectionMap;
  private marketConnMap: MarketStreamWsConnectionMap;
  private connectionMap: ConnectionMap;

  constructor(publisher: Publisher, marketSubscriber: MarketSubscriber, userConnectionMap: UserConnectionMap, marketConnMap: MarketStreamWsConnectionMap, connectionMap: ConnectionMap) {
    this.publisher = publisher;
    this.marketSubscriber = marketSubscriber;
    this.userConnectionMap = userConnectionMap;
    this.marketConnMap = marketConnMap;
    this.connectionMap = connectionMap;
  }

  public async Handler(message: ClientMessage, ws: WebSocket, connectionId: string) {

    switch (message.type) {
      case MessageType.enum.SUBSCRIBE_MARKET: {
        const { userID, marketID } = message.payload;
        if (userID) {
          this.userConnectionMap.AddUser(userID, ws);
        }
        this.marketConnMap.Add(marketID, ws);
        this.marketSubscriber.SubscribeMarket(marketID);

        await this.publisher.PublishMessageInMarket(marketID, JSON.stringify({
          MessageType: PUBSLISHED_MESSAGE_TYPES.orderBookSubs,
          UserId: userID,
          ConnectionId: connectionId,
        }));
        await this.publisher.PublishMessageInMarket(marketID, JSON.stringify({
          MessageType: PUBSLISHED_MESSAGE_TYPES.depthSubs,
          UserId: userID,
          ConnectionId: connectionId,
        }));

        if (userID) {
          await this.publisher.PublishMessageInMarket(marketID, JSON.stringify({
            MessageType: PUBSLISHED_MESSAGE_TYPES.walletLoad,
            UserId: userID,
            ConnectionId: connectionId,
          }));
          this.connectionMap.SetMeta(connectionId, userID, marketID);
        }

        ws.send(JSON.stringify({ type: "SUBSCRIBED", payload: { marketID } }));
        break;
      }
    }

  }

}


export default MessageHandler;
