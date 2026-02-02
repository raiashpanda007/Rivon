'use client';

import { Provider } from 'react-redux';
import { store } from './store';
import React, { useEffect } from 'react';
import Cookies from 'js-cookie';
import { setUserDetails, clearUserDetails } from './slices/userSlice';

export const StoreProvider = ({ children }: { children: React.ReactNode }) => {
    useEffect(() => {
        const syncAuthState = () => {
            const userCookie = Cookies.get("User_meta_data");
            const currentState = store.getState().user.userDetails;

            if (userCookie) {
                try {
                    const parsedUser = JSON.parse(userCookie);
                    if (!currentState || currentState.id !== parsedUser.id) {
                        store.dispatch(setUserDetails(parsedUser));
                    }
                } catch (e) {
                    console.error("Failed to parse auth cookie", e);
                }
            } else {
                if (currentState) {
                    store.dispatch(clearUserDetails());
                }
            }
        };

        syncAuthState();

        const interval = setInterval(syncAuthState, 1000);

        return () => clearInterval(interval);
    }, []);

    return <Provider store={store}>{children}</Provider>;
};
