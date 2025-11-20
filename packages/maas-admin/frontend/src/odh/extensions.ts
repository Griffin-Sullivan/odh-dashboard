import { DataScienceStackComponent } from '@odh-dashboard/internal/concepts/areas/types';
import type {
  NavExtension,
  RouteExtension,
  AreaExtension,
} from '@odh-dashboard/plugin-core/extension-points';

const PLUGIN_MAAS_ADMIN = 'plugin-maas-admin';

const extensions: (NavExtension | RouteExtension | AreaExtension)[] = [
  {
    type: 'app.area',
    properties: {
      id: PLUGIN_MAAS_ADMIN,
      requiredComponents: [DataScienceStackComponent.LLAMA_STACK_OPERATOR],
      featureFlags: ['modelAsService'],
    },
  },
  {
    type: 'app.navigation/href',
    flags: {
      required: [PLUGIN_MAAS_ADMIN],
    },
    properties: {
      id: 'tiers',
      title: 'Tiers',
      href: '/settings/tiers',
      section: 'settings',
      path: '/settings/tiers/*',
    },
  },
];

export default extensions;
