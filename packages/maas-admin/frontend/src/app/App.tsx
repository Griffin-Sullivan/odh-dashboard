import * as React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AppLayout } from '~/app/standalone/AppLayout';
import { AppRoutes } from '~/app/AppRoutes';
import '@patternfly/react-core/dist/styles/base.css';
import './app.css';

const App: React.FunctionComponent = () => (
  <AppLayout>
    <Routes>
      <Route path="/settings/tiers" element={<AppRoutes />} />
      <Route path="*" element={<Navigate to="/settings/tiers" replace />} />
    </Routes>
  </AppLayout>
);

export default App;
