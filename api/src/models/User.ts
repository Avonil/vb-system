import mongoose, { Schema, Document } from 'mongoose';

export interface IUser extends Document {
  twitchId: string;
  username: string;
  displayName: string;
  avatarUrl: string;
  
  isStreamer: boolean;
  isAdmin: boolean;

  auth: {
    accessToken: string;
    refreshToken: string;
    scopes: string[];
    expiresAt?: Date;
  };

  config: {
    language: string;
    prefix: string;
    botEnabled: boolean;
    timezone: string;
  };

  modules: {
    chat: {
      enabled: boolean;
      announceLive: boolean;
      customCommands: Array<{
        trigger: string;
        response: string;
        cooldown: number;
        userLevel: string;
        enabled: boolean;
      }>;
      visualCommands: Array<{
        trigger: string;
        logic: any[]; // Тут буде зберігатись JSON від Tauri
      }>;
    };
    moderation: {
      enabled: boolean;
      capsProtection: { active: boolean; limit: number; action: string };
      linkProtection: { active: boolean; permitClips: boolean };
      bannedWords: string[];
    };
    alerts: {
      enabled: boolean;
      discord: { enabled: boolean; serverId: string; channelId: string; message: string };
      telegram: { enabled: boolean; chatId: string; message: string };
      // Заділ на майбутні OBS HTML/JS віджети
      webOverlay: {
        followers: { active: boolean; html: string; css: string; js: string };
        donations: { active: boolean; minAmount: number; html: string; css: string; js: string };
        redemptions: { active: boolean }; // Бали каналу
      };
    };
  };

  meta: {
    lastLogin: Date;
    dashboardTheme: string;
    subscriptionTier: string;
  };
  
  createdAt: Date;
  updatedAt: Date;
}

const UserSchema = new Schema<IUser>({
  twitchId: { type: String, required: true, unique: true, index: true },
  username: { type: String, required: true, index: true },
  displayName: { type: String },
  avatarUrl: { type: String },

  isStreamer: { type: Boolean, default: false },
  isAdmin: { type: Boolean, default: false },

  auth: {
    accessToken: { type: String },
    refreshToken: { type: String },
    scopes: [{ type: String }],
    expiresAt: { type: Date }
  },

  config: {
    language: { type: String, default: 'ua' },
    prefix: { type: String, default: '!' },
    botEnabled: { type: Boolean, default: true },
    timezone: { type: String, default: 'Europe/Kyiv' }
  },

  modules: {
    chat: {
      enabled: { type: Boolean, default: true },
      announceLive: { type: Boolean, default: true },
      customCommands: [{
        trigger: String,
        response: String,
        cooldown: Number,
        userLevel: { type: String, default: 'everyone' },
        enabled: { type: Boolean, default: true }
      }],
      visualCommands: [{
        trigger: String,
        logic: [Schema.Types.Mixed] // Дозволяє зберігати будь-який JSON масив
      }]
    },
    moderation: {
      enabled: { type: Boolean, default: true },
      capsProtection: { active: { type: Boolean, default: false }, limit: { type: Number, default: 20 }, action: { type: String, default: 'timeout' } },
      linkProtection: { active: { type: Boolean, default: true }, permitClips: { type: Boolean, default: true } },
      bannedWords: [{ type: String }]
    },
    alerts: {
      enabled: { type: Boolean, default: false },
      discord: {
        enabled: { type: Boolean, default: false },
        serverId: { type: String, default: "" },
        channelId: { type: String, default: "" },
        message: { type: String, default: "@everyone Стрім почався!" }
      },
      telegram: {
        enabled: { type: Boolean, default: false },
        chatId: { type: String, default: "" },
        message: { type: String, default: "Стрім почався!" }
      },
      webOverlay: {
        followers: { active: { type: Boolean, default: true }, html: { type: String, default: "" }, css: { type: String, default: "" }, js: { type: String, default: "" } },
        donations: { active: { type: Boolean, default: false }, minAmount: { type: Number, default: 10 }, html: { type: String, default: "" }, css: { type: String, default: "" }, js: { type: String, default: "" } },
        redemptions: { active: { type: Boolean, default: true } }
      }
    }
  },

  meta: {
    lastLogin: { type: Date, default: Date.now },
    dashboardTheme: { type: String, default: 'dark_cyberpunk' },
    subscriptionTier: { type: String, default: 'free' }
  }
}, { timestamps: true });

export const User = mongoose.model<IUser>('User', UserSchema);