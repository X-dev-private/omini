// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)

package transfer

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	ibctransfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/omini/omini/v20/x/ibc/transfer/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic embeds the IBC Transfer AppModuleBasic
type AppModuleBasic struct {
	*ibctransfer.AppModuleBasic
}

// AppModule represents the AppModule for this module
type AppModule struct {
	*ibctransfer.AppModule
	keeper keeper.Keeper
}

// NewAppModule creates a new 20-transfer module
func NewAppModule(k keeper.Keeper) AppModule {
	am := ibctransfer.NewAppModule(*k.Keeper)
	return AppModule{
		AppModule: &am,
		keeper:    k,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Override Transfer Msg Server
	types.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := ibctransferkeeper.NewMigrator(*am.keeper.Keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.MigrateTraces); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 2, m.MigrateTotalEscrowForDenom); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 3, m.MigrateParams); err != nil {
		panic(fmt.Errorf("failed to migrate transfer app version 3 to 4 (self-managed params migration): %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 4, m.MigrateDenomMetadata); err != nil {
		panic(fmt.Errorf("failed to migrate transfer app from version 4 to 5 (set denom metadata migration): %v", err))
	}
}
