import * as React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { NotFound } from 'mod-arch-shared';
import { NavDataItem } from '~/app/standalone/types';

import '@patternfly/chatbot/dist/css/main.css';
import TiersPage from './TiersPage';

export interface IAppRoute {
  label?: string; // Excluding the label will exclude the route from the nav sidebar in AppLayout
  element: React.ReactElement;
  exact?: boolean;
  path: string;
  title: string;
  routes?: undefined;
}

export interface IAppRouteGroup {
  label: string;
  routes: IAppRoute[];
}

export type AppRouteConfig = IAppRoute | IAppRouteGroup;

export const useNavData = (): NavDataItem[] => [
  {
    label: 'Tiers',
    children: [
      {
        label: 'Tiers',
        path: '/settings/tiers',
        href: '/settings/tiers',
      },
    ],
  },
];

const AppRoutes = (): React.ReactElement => (
  <Routes>
    <Route path="/" element={<Navigate to="/settings/tiers" replace />} />
    <Route path="/settings/tiers" element={<TiersPage />} />
    <Route path="*" element={<NotFound />} />
  </Routes>
);

export { AppRoutes };
