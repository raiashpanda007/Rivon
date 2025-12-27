'use client';

import { Provider } from 'react-redux';
import { store } from './store.js';
import React from 'react';

export const StoreProvider = ({ children }: { children: React.ReactNode }) => {
    return <Provider store={store}>{children}</Provider>;
};
