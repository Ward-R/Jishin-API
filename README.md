# 地震 API (Jishin API) 🌋

**日本語の README は英語版の後に続きます。**  
*Japanese language README follows the English version.*

---

# Jishin API - Real-time Earthquake Data from Japan 🇯🇵

A production-ready REST API that provides real-time earthquake data from the Japan Meteorological Agency (JMA). Built with Go, deployed on AWS Lambda, and designed for high performance and reliability.

## 🚀 Live API

**Base URL**: `https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod`

Try it now: [API Documentation](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/)

## 📊 Features

- **Real-time Data**: Automatically syncs with JMA earthquake reports
- **Bilingual Support**: All earthquake data includes both Japanese and English descriptions
- **High Performance**: Sub-50ms response times via AWS Lambda
- **Scalable Architecture**: Serverless design handles traffic spikes automatically
- **Cost Effective**: Built to operate within AWS free tier limits

## 🔗 API Endpoints

| Endpoint | Method | Description | Example |
|----------|---------|------------|---------|
| `/` | GET | API documentation | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/) |
| `/health` | GET | Health check & database status | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/health) |
| `/earthquakes` | GET | Latest 50 earthquakes | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes) |
| `/earthquakes?limit=10` | GET | Limit results | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes?limit=10) |
| `/earthquakes?magnitude=5.0` | GET | Filter by magnitude | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes?magnitude=5.0) |
| `/earthquakes/stats` | GET | Database statistics | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/stats) |
| `/earthquakes/recent` | GET | Last 24 hours | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/recent) |
| `/earthquakes/largest/today` | GET | Strongest earthquake today | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/largest/today) |
| `/earthquakes/largest/week` | GET | Strongest this week | [Try it](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/largest/week) |
| `/earthquake/{id}` | GET | Specific earthquake by ID | Example: `/earthquake/20250812113450` |
| `/sync` | POST | Manual data sync (admin) | Triggers JMA data update |

## 🏗️ Architecture

```
JMA Data Source → AWS Lambda (Go) → Supabase PostgreSQL → API Gateway → Public API
```

- **Backend**: Go with clean architecture (handlers, services, database layers)
- **Database**: Supabase PostgreSQL with automatic connection pooling
- **Hosting**: AWS Lambda with API Gateway (Tokyo region: `ap-northeast-1`)
- **Data Source**: Japan Meteorological Agency official earthquake reports

## 📈 Sample Response

```json
{
  "ReportId": "20250812113450",
  "OriginTime": "2025-08-12T02:34:00Z",
  "Magnitude": 2.7,
  "DepthKm": 10,
  "Latitude": 29.3,
  "Longitude": 129.5,
  "MaxIntensity": "1",
  "JpLocation": "トカラ列島近海",
  "EnLocation": "Adjacent Sea of Tokara Islands",
  "JpComment": "この地震による津波の心配はありません。",
  "EnComment": "This earthquake poses no tsunami risk."
}
```

## 🛠️ Technology Stack

- **Language**: Go 1.24
- **Cloud Provider**: AWS (Lambda + API Gateway)
- **Database**: Supabase PostgreSQL
- **Architecture**: Serverless with clean separation of concerns
- **Deployment**: Automated via AWS Console
- **Monitoring**: CloudWatch logs and metrics

## 📦 Project Structure

```
Jishin-API/
├── api/          # HTTP handlers and request/response logic
├── db/           # Database queries and connection management
├── service/      # Business logic and external API calls
├── types/        # Data structures and models
├── main.go       # Lambda entry point and routing
└── README.md     # This file
```

## 🎯 Development Philosophy

This project demonstrates professional software development practices:

- **Clean Architecture**: Separation of concerns with dedicated packages
- **Error Handling**: Comprehensive error management and logging
- **Documentation**: Self-documenting API with clear endpoint descriptions
- **Testing**: Built-in health checks and monitoring
- **Security**: Environment variable management and secure database connections
- **Performance**: Optimized for serverless cold starts and high throughput

## 🌟 Learning Journey

This project was developed as an original work to learn modern Go development and AWS cloud architecture. Development was assisted by Claude AI for learning Go best practices, AWS deployment patterns, and professional coding standards. The core logic, architecture decisions, and implementation approach reflect personal learning and problem-solving.

## 🌐 Live Implementation

This API is actively used on my portfolio website at [www.ryanward.dev](https://www.ryanward.dev) to display real-time earthquake data and demonstrate full-stack integration capabilities.

---

# 地震 API - 日本からのリアルタイム地震データ 🇯🇵

日本気象庁（JMA）からのリアルタイム地震データを提供する本格的なREST API。Go言語で構築され、AWS Lambdaにデプロイされ、高性能と信頼性を重視して設計されています。

## 🚀 ライブAPI

**ベースURL**: `https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod`

今すぐ試してください: [APIドキュメント](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/)

## 📊 機能

- **リアルタイムデータ**: JMAの地震報告と自動同期
- **二カ国語対応**: すべての地震データに日本語と英語の説明を含む
- **高性能**: AWS Lambda経由で50ms未満の応答時間
- **スケーラブルアーキテクチャ**: サーバーレス設計でトラフィックスパイクを自動処理
- **コスト効率**: AWS無料利用枠内での運用を想定した設計

## 🔗 APIエンドポイント

| エンドポイント | メソッド | 説明 | 例 |
|----------|---------|------------|---------|
| `/` | GET | APIドキュメント | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/) |
| `/health` | GET | ヘルスチェックとデータベース状態 | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/health) |
| `/earthquakes` | GET | 最新50件の地震 | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes) |
| `/earthquakes?limit=10` | GET | 結果を制限 | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes?limit=10) |
| `/earthquakes?magnitude=5.0` | GET | マグニチュードでフィルター | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes?magnitude=5.0) |
| `/earthquakes/stats` | GET | データベース統計 | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/stats) |
| `/earthquakes/recent` | GET | 過去24時間 | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/recent) |
| `/earthquakes/largest/today` | GET | 今日の最大地震 | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/largest/today) |
| `/earthquakes/largest/week` | GET | 今週の最大地震 | [試してみる](https://aftbll7aci.execute-api.ap-northeast-1.amazonaws.com/prod/earthquakes/largest/week) |
| `/earthquake/{id}` | GET | IDによる特定の地震 | 例: `/earthquake/20250812113450` |
| `/sync` | POST | 手動データ同期（管理者用） | JMAデータ更新をトリガー |

## 🏗️ アーキテクチャ

```
JMAデータソース → AWS Lambda (Go) → Supabase PostgreSQL → API Gateway → パブリックAPI
```

- **バックエンド**: クリーンアーキテクチャのGo（ハンドラー、サービス、データベース層）
- **データベース**: 自動接続プーリング付きSupabase PostgreSQL
- **ホスティング**: API Gateway付きAWS Lambda（東京リージョン: `ap-northeast-1`）
- **データソース**: 気象庁公式地震報告

## 🛠️ 技術スタック

- **言語**: Go 1.24
- **クラウドプロバイダー**: AWS（Lambda + API Gateway）
- **データベース**: Supabase PostgreSQL
- **アーキテクチャ**: 関心事の分離を持つサーバーレス
- **デプロイメント**: AWSコンソール経由で自動化
- **監視**: CloudWatchログとメトリクス

## 🌟 学習の旅

このプロジェクトは、現代のGo開発とAWSクラウドアーキテクチャを学ぶためのオリジナル作品として開発されました。Goのベストプラクティス、AWSデプロイメントパターン、専門的なコーディング標準を学ぶために、Claude AIのサポートを受けて開発されました。コアロジック、アーキテクチャの決定、実装アプローチは個人的な学習と問題解決を反映しています。

## 🌐 ライブ実装

このAPIは私のポートフォリオサイト [www.ryanward.dev](https://www.ryanward.dev) でリアルタイム地震データの表示に使用され、フルスタック統合機能を実証しています。

---

**Built with ❤️ in Tokyo | 東京で❤️を込めて構築**