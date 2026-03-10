import mongoose from 'mongoose';

export const connectDB = async () => {
  try {
    const uri = process.env.MONGO_URI;
    if (!uri) throw new Error('❌ MONGO_URI is missing in .env');

    await mongoose.connect(uri);
    
    console.log('🍃 MongoDB Connected!');
  } catch (error) {
    console.error('❌ MongoDB Connection Failed:', error);
    process.exit(1);
  }
};