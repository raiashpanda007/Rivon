import type WebSocket from "ws";
import { z as zod } from "zod";

export const MessageType = zod.enum([
  "SUBSCRIBE_MARKET",
  "UNSUBSCRIBE_MARKET",
]);

const ClientMsgSchema = zod.discriminatedUnion("type", [
  zod.object({
    type: zod.literal("SUBSCRIBE_MARKET"),
    payload: zod.object({
      marketID: zod.string().uuid(), // stricter validation
      userID: zod.string().optional(),
    }),
  }),

  zod.object({
    type: zod.literal("UNSUBSCRIBE_MARKET"),
    payload: zod.object({
      marketID: zod.string().uuid(),
    }),
  }),
]);

export type ClientMessage = zod.infer<typeof ClientMsgSchema>;

export function ParseClient(msg: WebSocket.RawData): ClientMessage | null {
  try {
    const parsed = JSON.parse(msg.toString());

    const resp = ClientMsgSchema.safeParse(parsed);

    if (!resp.success) {
      console.error("Invalid message format:", resp.error.format());
      return null;
    }

    return resp.data;
  } catch (e) {
    console.error("Invalid JSON:", e);
    return null;
  }
}

const InvalidJSON = {
  type: "ERROR",
  payload: {
    message: "Invalid JSON. Unable to parse JSON",
  },
};

const Types = {
  ErrorMessage: {
    InvalidJSON,
  },
};


export enum PUBSLISHED_MESSAGE_TYPES {
  orderBookSubs = "ORDER_BOOK_SUBSCRIBE",
  depthSubs = "DEPTH_SUBSCRIBE",
  walletLoad = "WALLET_LOAD",
  walletEvict = "WALLET_EVICT",
}


export default Types;
