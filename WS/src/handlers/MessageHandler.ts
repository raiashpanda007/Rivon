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

        this.marketConnMap.Add(marketID, ws);
        this.marketSubscriber.SubscribeMarket(marketID);

        // Always store marketId so cleanup fires for all connections, not just authed ones
        this.connectionMap.SetMeta(connectionId, marketID, userID);

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
          this.userConnectionMap.AddUser(userID, ws);
          await this.publisher.PublishMessageInMarket(marketID, JSON.stringify({
            MessageType: PUBSLISHED_MESSAGE_TYPES.walletLoad,
            UserId: userID,
            ConnectionId: connectionId,
          }));
        }

        ws.send(JSON.stringify({ type: "SUBSCRIBED", payload: { marketID } }));
        break;
      }

      case MessageType.enum.UNSUBSCRIBE_MARKET: {
        const { marketID } = message.payload;
        this.marketConnMap.Remove(marketID, ws);
        ws.send(JSON.stringify({ type: "UNSUBSCRIBED", payload: { marketID } }));
        break;
      }

      case MessageType.enum.CANCEL_ORDER: {
        const { marketID, orderId, cancelQty } = message.payload;
        const meta = this.connectionMap.GetMeta(connectionId);
        if (!meta?.userId) {
          ws.send(JSON.stringify({ type: "ERROR", payload: { message: "Not authenticated" } }));
          break;
        }
        await this.publisher.PublishMessageInMarket(marketID, JSON.stringify({
          MessageType: "CANCEL_ORDER_WS",
          UserId: meta.userId,
          OrderId: orderId,
          CancelQty: cancelQty ?? 0,
          ConnectionId: connectionId,
        }));
        break;
      }
    }

  }

}


export default MessageHandler;
