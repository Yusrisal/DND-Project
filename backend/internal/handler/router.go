package handler

import (
	"database/sql"
	"net/http"

	"dndproject/backend/internal/config"
	"dndproject/backend/internal/middleware"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
	"dndproject/backend/internal/ws"
)

func health(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, `{"status":"unhealthy"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

func NewRouter(cfg config.Config, db *sql.DB, hub *ws.Hub) http.Handler {
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	characterRepo := repository.NewCharacterRepository(db)
	monsterRepo := repository.NewMonsterRepository(db)
	shopRepo := repository.NewShopRepository(db)
	compendiumRepo := repository.NewCompendiumRepository(db)
	handoutRepo := repository.NewHandoutRepository(db)
	savepointRepo := repository.NewSavepointRepository(db)
	encounterRepo := repository.NewEncounterRepository(db)
	skillRepo := repository.NewSkillRepository(db)

	authService := service.NewAuthService(userRepo, repository.NewSessionRepository(db))
	mapService := service.NewMapService(repository.NewMapRepository(db))
	tokenService := service.NewTokenService(tokenRepo, characterRepo)
	chatService := service.NewChatService(repository.NewChatRepository(db))
	rollService := service.NewRollService(repository.NewRollRepository(db))
	characterService := service.NewCharacterService(characterRepo, skillRepo)
	skillService := service.NewSkillService(skillRepo, characterRepo)
	monsterService := service.NewMonsterService(monsterRepo, tokenRepo)
	shopService := service.NewShopService(shopRepo, characterRepo)
	compendiumService := service.NewCompendiumService(compendiumRepo, monsterRepo)
	handoutService := service.NewHandoutService(handoutRepo)
	savepointService := service.NewSavepointService(savepointRepo, tokenRepo, characterRepo, encounterRepo)
	combatService := service.NewCombatService(
		encounterRepo, tokenRepo, characterRepo, monsterRepo, characterService, rollService)

	g := &guard{auth: authService, secure: cfg.Env == "production"}

	campaignService := service.NewCampaignService(repository.NewCampaignRepository(db))
	auth := NewAuthHandler(authService, g)
	invites := NewInviteHandler(service.NewInviteService(repository.NewInviteRepository(db)))
	campaigns := NewCampaignHandler(campaignService)
	maps := NewMapHandler(mapService)
	tokens := NewTokenHandler(tokenService, mapService, hub)
	chat := NewChatHandler(chatService)
	characters := NewCharacterHandler(characterService)
	monsters := NewMonsterHandler(monsterService, mapService, hub)
	combat := NewCombatHandler(combatService, mapService, hub)
	shops := NewShopHandler(shopService)
	compendium := NewCompendiumHandler(compendiumService)
	handouts := NewHandoutHandler(handoutService)
	files := NewUploadHandler(cfg.UploadDir)
	savepoints := NewSavepointHandler(savepointService)
	skills := NewSkillHandler(skillService, characterService)
	sockets := NewWSHandler(
		tokenService, mapService, campaignService, chatService, rollService, characterService, combatService,
		authService, hub, cfg.CORSOrigins)

	mux := http.NewServeMux()

	// Public.
	mux.HandleFunc("GET /api/health", health(db))
	mux.HandleFunc("POST /api/auth/register", auth.Register)
	mux.HandleFunc("POST /api/auth/login", auth.Login)
	mux.HandleFunc("POST /api/auth/logout", auth.Logout)

	// Any signed-in account: reading the world.
	mux.HandleFunc("GET /api/auth/me", g.user(auth.Me))
	mux.HandleFunc("GET /api/users", g.user(auth.Users))
	mux.HandleFunc("GET /api/campaigns", g.user(campaigns.List))
	mux.HandleFunc("GET /api/campaigns/{id}", g.user(campaigns.Get))
	mux.HandleFunc("GET /api/campaigns/{id}/handouts", g.user(handouts.List))
	mux.HandleFunc("GET /api/handouts/{id}", g.user(handouts.Get))
	mux.HandleFunc("GET /api/maps", g.user(maps.List))
	mux.HandleFunc("GET /api/maps/{id}", g.user(maps.Get))
	mux.HandleFunc("GET /api/maps/{id}/tokens", g.user(tokens.Board))
	mux.HandleFunc("GET /api/maps/{id}/encounter", g.user(combat.Encounter))
	mux.HandleFunc("GET /api/maps/{id}/aoe-templates", g.user(maps.ListAoETemplates))
	mux.HandleFunc("GET /api/maps/{id}/objects", g.user(maps.ListObjects))
	mux.HandleFunc("PUT /api/map-objects/{id}", g.user(maps.UpdateObject))
	mux.HandleFunc("GET /api/tokens", g.user(tokens.List))
	mux.HandleFunc("GET /api/tokens/{id}", g.user(tokens.Get))
	mux.HandleFunc("GET /api/chat", g.user(chat.List))
	mux.HandleFunc("GET /api/monsters", g.user(monsters.List))
	mux.HandleFunc("GET /api/monsters/{id}", g.user(monsters.StatBlock))
	mux.HandleFunc("GET /api/shops", g.user(shops.List))
	mux.HandleFunc("GET /api/shops/{id}", g.user(shops.Get))
	mux.HandleFunc("POST /api/shop-items/{id}/buy", g.user(shops.BuyItem))
	mux.HandleFunc("GET /api/environments", g.user(compendium.ListEnvironments))
	mux.HandleFunc("GET /api/environments/{id}", g.user(compendium.GetEnvironment))
	mux.HandleFunc("GET /api/compendium/items", g.user(compendium.ListItems))
	mux.HandleFunc("GET /api/compendium/items/{id}", g.user(compendium.GetItem))
	mux.HandleFunc("GET /api/compendium/spells", g.user(compendium.ListSpells))
	mux.HandleFunc("GET /api/compendium/spells/{id}", g.user(compendium.GetSpell))
	mux.HandleFunc("GET /api/status-effects", g.user(compendium.ListStatusEffects))
	mux.HandleFunc("GET /api/status-effects/{id}", g.user(compendium.GetStatusEffect))
	mux.HandleFunc("GET /api/monsters/{id}/drops", g.user(compendium.ListMonsterDrops))
	mux.HandleFunc("GET /uploads/", g.user(uploads(cfg.UploadDir).ServeHTTP))
	mux.HandleFunc("GET /ws", sockets.Serve)

	// Character sheets
	mux.HandleFunc("GET /api/characters", g.user(characters.List))
	mux.HandleFunc("POST /api/characters", g.user(characters.Create))
	mux.HandleFunc("GET /api/characters/{id}", g.user(characters.Sheet))
	mux.HandleFunc("PUT /api/characters/{id}", g.user(characters.Update))
	mux.HandleFunc("DELETE /api/characters/{id}", g.user(characters.Delete))
	mux.HandleFunc("POST /api/characters/{id}/long-rest", g.user(characters.LongRest))
	mux.HandleFunc("POST /api/characters/{id}/short-rest", g.user(characters.ShortRest))
	mux.HandleFunc("POST /api/characters/{id}/death-save", g.user(characters.DeathSave))
	mux.HandleFunc("POST /api/characters/{id}/items", g.user(characters.AddItem))
	mux.HandleFunc("POST /api/characters/{id}/spells", g.user(characters.AddSpell))
	mux.HandleFunc("PUT /api/characters/{id}/slots/{level}", g.user(characters.SetSlot))
	mux.HandleFunc("PUT /api/characters/{id}/proficiencies", g.user(characters.SetProficiency))
	mux.HandleFunc("PUT /api/items/{id}", g.user(characters.UpdateItem))
	mux.HandleFunc("DELETE /api/items/{id}", g.user(characters.DeleteItem))
	mux.HandleFunc("POST /api/characters/{id}/transfer-item", g.user(skills.TransferItem))
	mux.HandleFunc("POST /api/characters/{id}/allocate-stat", g.user(skills.AllocateStat))
	mux.HandleFunc("GET /api/skill-trees", g.user(skills.ListTrees))
	mux.HandleFunc("POST /api/characters/{id}/learn-skill", g.user(skills.Learn))
	mux.HandleFunc("POST /api/characters/{id}/toggle-skill", g.user(skills.ToggleActive))
	mux.HandleFunc("POST /api/characters/{id}/use-scroll", g.user(skills.UseScroll))
	mux.HandleFunc("PUT /api/spells/{id}", g.user(characters.UpdateSpell))
	mux.HandleFunc("DELETE /api/spells/{id}", g.user(characters.DeleteSpell))

	// DM Only
	mux.HandleFunc("GET /api/invites", g.dm(invites.List))
	mux.HandleFunc("POST /api/invites", g.dm(invites.Create))
	mux.HandleFunc("DELETE /api/invites/{code}", g.dm(invites.Delete))
	mux.HandleFunc("POST /api/campaigns", g.dm(campaigns.Create))
	mux.HandleFunc("PUT /api/campaigns/{id}", g.dm(campaigns.Update))
	mux.HandleFunc("PUT /api/campaigns/{id}/pause", g.dm(campaigns.SetPause))
	mux.HandleFunc("PUT /api/campaigns/{id}/retention", g.dm(campaigns.SetRetention))
	mux.HandleFunc("DELETE /api/campaigns/{id}", g.dm(campaigns.Delete))
	mux.HandleFunc("GET /api/campaigns/{id}/savepoints", g.dm(savepoints.List))
	mux.HandleFunc("POST /api/campaigns/{id}/savepoints", g.dm(savepoints.Create))
	mux.HandleFunc("POST /api/savepoints/{id}/restore", g.dm(savepoints.Restore))
	mux.HandleFunc("DELETE /api/savepoints/{id}", g.dm(savepoints.Delete))
	mux.HandleFunc("POST /api/characters/{id}/xp", g.dm(skills.AddXP))
	mux.HandleFunc("PUT /api/characters/{id}/level-block", g.dm(skills.SetLevelBlock))
	mux.HandleFunc("POST /api/characters/{id}/grant-points", g.dm(skills.GrantPoints))
	mux.HandleFunc("GET /api/pending-loot", g.dm(skills.ListPendingLoot))
	mux.HandleFunc("POST /api/pending-loot/{id}/resolve", g.dm(skills.ResolvePendingLoot))
	mux.HandleFunc("POST /api/uploads", g.dm(files.Create))
	mux.HandleFunc("POST /api/maps", g.dm(maps.Create))
	mux.HandleFunc("PUT /api/maps/{id}", g.dm(maps.Update))
	mux.HandleFunc("PUT /api/maps/{id}/fog", g.dm(maps.UpdateFog))
	mux.HandleFunc("PUT /api/maps/{id}/weather", g.dm(maps.SetWeather))
	mux.HandleFunc("DELETE /api/maps/{id}", g.dm(maps.Delete))
	mux.HandleFunc("POST /api/maps/{id}/aoe-templates", g.dm(maps.CreateAoETemplate))
	mux.HandleFunc("DELETE /api/aoe-templates/{id}", g.dm(maps.DeleteAoETemplate))
	mux.HandleFunc("POST /api/maps/{id}/objects", g.dm(maps.CreateObject))
	mux.HandleFunc("DELETE /api/map-objects/{id}", g.dm(maps.DeleteObject))
	mux.HandleFunc("POST /api/campaigns/{id}/handouts", g.dm(handouts.Create))
	mux.HandleFunc("PUT /api/handouts/{id}", g.dm(handouts.Update))
	mux.HandleFunc("DELETE /api/handouts/{id}", g.dm(handouts.Delete))
	mux.HandleFunc("POST /api/tokens", g.dm(tokens.Create))
	mux.HandleFunc("PUT /api/tokens/{id}", g.dm(tokens.Update))
	mux.HandleFunc("DELETE /api/tokens/{id}", g.dm(tokens.Delete))
	mux.HandleFunc("PUT /api/tokens/{id}/conditions", g.dm(combat.SetCondition))

	// Shops management & loot split
	mux.HandleFunc("POST /api/shops", g.dm(shops.Create))
	mux.HandleFunc("PUT /api/shops/{id}", g.dm(shops.Update))
	mux.HandleFunc("DELETE /api/shops/{id}", g.dm(shops.Delete))
	mux.HandleFunc("POST /api/shops/{id}/items", g.dm(shops.AddItem))
	mux.HandleFunc("PUT /api/shop-items/{id}", g.dm(shops.UpdateItem))
	mux.HandleFunc("DELETE /api/shop-items/{id}", g.dm(shops.DeleteItem))
	mux.HandleFunc("POST /api/loot/split", g.dm(characters.SplitLoot))

	// Environments & Compendium
	mux.HandleFunc("POST /api/environments", g.dm(compendium.CreateEnvironment))
	mux.HandleFunc("PUT /api/environments/{id}", g.dm(compendium.UpdateEnvironment))
	mux.HandleFunc("DELETE /api/environments/{id}", g.dm(compendium.DeleteEnvironment))
	mux.HandleFunc("POST /api/compendium/items", g.dm(compendium.CreateItem))
	mux.HandleFunc("PUT /api/compendium/items/{id}", g.dm(compendium.UpdateItem))
	mux.HandleFunc("DELETE /api/compendium/items/{id}", g.dm(compendium.DeleteItem))
	mux.HandleFunc("POST /api/monsters/{id}/drops", g.dm(compendium.CreateMonsterDrop))
	mux.HandleFunc("DELETE /api/monster-drops/{id}", g.dm(compendium.DeleteMonsterDrop))
	mux.HandleFunc("POST /api/monsters/{id}/roll-loot", g.dm(compendium.RollMonsterLoot))
	mux.HandleFunc("POST /api/compendium/spells", g.dm(compendium.CreateSpell))
	mux.HandleFunc("PUT /api/compendium/spells/{id}", g.dm(compendium.UpdateSpell))
	mux.HandleFunc("DELETE /api/compendium/spells/{id}", g.dm(compendium.DeleteSpell))
	mux.HandleFunc("POST /api/status-effects", g.dm(compendium.CreateStatusEffect))
	mux.HandleFunc("PUT /api/status-effects/{id}", g.dm(compendium.UpdateStatusEffect))
	mux.HandleFunc("DELETE /api/status-effects/{id}", g.dm(compendium.DeleteStatusEffect))

	// Monsters and encounters
	mux.HandleFunc("POST /api/monsters", g.dm(monsters.Create))
	mux.HandleFunc("PUT /api/monsters/{id}", g.dm(monsters.Update))
	mux.HandleFunc("DELETE /api/monsters/{id}", g.dm(monsters.Delete))
	mux.HandleFunc("PUT /api/monsters/{id}/proficiencies", g.dm(monsters.SetProficiency))
	mux.HandleFunc("POST /api/monsters/{id}/actions", g.dm(monsters.AddAction))
	mux.HandleFunc("PUT /api/actions/{id}", g.dm(monsters.UpdateAction))
	mux.HandleFunc("DELETE /api/actions/{id}", g.dm(monsters.DeleteAction))
	mux.HandleFunc("POST /api/spawn", g.dm(monsters.Spawn))

	mux.HandleFunc("POST /api/maps/{id}/encounter", g.dm(combat.Start))
	mux.HandleFunc("DELETE /api/maps/{id}/encounter", g.dm(combat.End))
	mux.HandleFunc("POST /api/maps/{id}/encounter/turn", g.dm(combat.NextTurn))
	mux.HandleFunc("POST /api/maps/{id}/encounter/combatants", g.dm(combat.AddCombatant))
	mux.HandleFunc("PUT /api/maps/{id}/encounter/combatants/{combatant}", g.dm(combat.SetInitiative))

	// In production, serve the built frontend from WebDir for all non-API routes.
	if cfg.WebDir != "" {
		mux.Handle("/", spa(cfg.WebDir))
	}

	var root http.Handler = mux
	root = middleware.Logger(root)
	root = middleware.Recover(root)
	root = middleware.CORS(cfg.CORSOrigins)(root)
	return root
}
