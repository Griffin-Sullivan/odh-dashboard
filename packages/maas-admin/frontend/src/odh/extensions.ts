import { DataScienceStackComponent } from '@odh-dashboard/internal/concepts/areas/types';
import type { NavExtension, AreaExtension } from '@odh-dashboard/plugin-core/extension-points';

const PLUGIN_MAAS_ADMIN = 'plugin-maas-admin';

const extensions: (NavExtension | AreaExtension)[] = [
  {
    type: 'app.area',
    properties: {
      id: PLUGIN_MAAS_ADMIN,
      reliantAreas: [PLUGIN_MAAS_ADMIN],
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
      href: '/maas-admin/tiers',
      section: 'settings',
      path: '/maas-admin/tiers/*',
      label: 'Tech Preview',
    },
  },
];

export default extensions;
