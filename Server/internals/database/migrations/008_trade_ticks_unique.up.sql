ALTER TABLE trade_ticks
    ADD CONSTRAINT trade_ticks_trade_id_unique UNIQUE (time, trade_id);
