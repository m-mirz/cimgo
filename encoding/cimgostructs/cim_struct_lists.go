package cimgostructs

import "fmt"

type CIMElementList struct {
	Elements map[string]interface{}
    ACDCConverterDCTerminals map[string]*ACDCConverterDCTerminal
    ACLineSegments map[string]*ACLineSegment
    Accumulators map[string]*Accumulator
    AccumulatorLimits map[string]*AccumulatorLimit
    AccumulatorLimitSets map[string]*AccumulatorLimitSet
    AccumulatorResets map[string]*AccumulatorReset
    AccumulatorValues map[string]*AccumulatorValue
    ActivePowerLimits map[string]*ActivePowerLimit
    Analogs map[string]*Analog
    AnalogLimits map[string]*AnalogLimit
    AnalogLimitSets map[string]*AnalogLimitSet
    AnalogValues map[string]*AnalogValue
    ApparentPowerLimits map[string]*ApparentPowerLimit
    AsynchronousMachines map[string]*AsynchronousMachine
    AsynchronousMachineEquivalentCircuits map[string]*AsynchronousMachineEquivalentCircuit
    AsynchronousMachineTimeConstantReactances map[string]*AsynchronousMachineTimeConstantReactance
    AsynchronousMachineUserDefineds map[string]*AsynchronousMachineUserDefined
    BaseVoltages map[string]*BaseVoltage
    BatteryUnits map[string]*BatteryUnit
    Bays map[string]*Bay
    BoundaryPoints map[string]*BoundaryPoint
    Breakers map[string]*Breaker
    BusNameMarkers map[string]*BusNameMarker
    BusbarSections map[string]*BusbarSection
    CAESPlants map[string]*CAESPlant
    CSCUserDefineds map[string]*CSCUserDefined
    Clamps map[string]*Clamp
    CogenerationPlants map[string]*CogenerationPlant
    CombinedCyclePlants map[string]*CombinedCyclePlant
    Commands map[string]*Command
    ConformLoads map[string]*ConformLoad
    ConformLoadGroups map[string]*ConformLoadGroup
    ConformLoadSchedules map[string]*ConformLoadSchedule
    ConnectivityNodes map[string]*ConnectivityNode
    ControlAreas map[string]*ControlArea
    ControlAreaGeneratingUnits map[string]*ControlAreaGeneratingUnit
    CoordinateSystems map[string]*CoordinateSystem
    CsConverters map[string]*CsConverter
    CurrentLimits map[string]*CurrentLimit
    CurrentTransformers map[string]*CurrentTransformer
    CurveDatas map[string]*CurveData
    Cuts map[string]*Cut
    DCBreakers map[string]*DCBreaker
    DCBusbars map[string]*DCBusbar
    DCChoppers map[string]*DCChopper
    DCConverterUnits map[string]*DCConverterUnit
    DCDisconnectors map[string]*DCDisconnector
    DCGrounds map[string]*DCGround
    DCLines map[string]*DCLine
    DCLineSegments map[string]*DCLineSegment
    DCNodes map[string]*DCNode
    DCSeriesDevices map[string]*DCSeriesDevice
    DCShunts map[string]*DCShunt
    DCSwitchs map[string]*DCSwitch
    DCTerminals map[string]*DCTerminal
    DCTopologicalIslands map[string]*DCTopologicalIsland
    DCTopologicalNodes map[string]*DCTopologicalNode
    DayTypes map[string]*DayType
    Diagrams map[string]*Diagram
    DiagramObjects map[string]*DiagramObject
    DiagramObjectGluePoints map[string]*DiagramObjectGluePoint
    DiagramObjectPoints map[string]*DiagramObjectPoint
    DiagramObjectStyles map[string]*DiagramObjectStyle
    DiagramStyles map[string]*DiagramStyle
    DifferenceModels map[string]*DifferenceModel
    DiscExcContIEEEDEC1As map[string]*DiscExcContIEEEDEC1A
    DiscExcContIEEEDEC2As map[string]*DiscExcContIEEEDEC2A
    DiscExcContIEEEDEC3As map[string]*DiscExcContIEEEDEC3A
    DisconnectingCircuitBreakers map[string]*DisconnectingCircuitBreaker
    Disconnectors map[string]*Disconnector
    DiscontinuousExcitationControlUserDefineds map[string]*DiscontinuousExcitationControlUserDefined
    Discretes map[string]*Discrete
    DiscreteValues map[string]*DiscreteValue
    EnergyConsumers map[string]*EnergyConsumer
    EnergySchedulingTypes map[string]*EnergySchedulingType
    EnergySources map[string]*EnergySource
    Equipments map[string]*Equipment
    EquivalentBranchs map[string]*EquivalentBranch
    EquivalentInjections map[string]*EquivalentInjection
    EquivalentNetworks map[string]*EquivalentNetwork
    EquivalentShunts map[string]*EquivalentShunt
    ExcAC1As map[string]*ExcAC1A
    ExcAC2As map[string]*ExcAC2A
    ExcAC3As map[string]*ExcAC3A
    ExcAC4As map[string]*ExcAC4A
    ExcAC5As map[string]*ExcAC5A
    ExcAC6As map[string]*ExcAC6A
    ExcAC8Bs map[string]*ExcAC8B
    ExcANSs map[string]*ExcANS
    ExcAVR1s map[string]*ExcAVR1
    ExcAVR2s map[string]*ExcAVR2
    ExcAVR3s map[string]*ExcAVR3
    ExcAVR4s map[string]*ExcAVR4
    ExcAVR5s map[string]*ExcAVR5
    ExcAVR7s map[string]*ExcAVR7
    ExcBBCs map[string]*ExcBBC
    ExcCZs map[string]*ExcCZ
    ExcDC1As map[string]*ExcDC1A
    ExcDC2As map[string]*ExcDC2A
    ExcDC3As map[string]*ExcDC3A
    ExcDC3A1s map[string]*ExcDC3A1
    ExcELIN1s map[string]*ExcELIN1
    ExcELIN2s map[string]*ExcELIN2
    ExcHUs map[string]*ExcHU
    ExcIEEEAC1As map[string]*ExcIEEEAC1A
    ExcIEEEAC2As map[string]*ExcIEEEAC2A
    ExcIEEEAC3As map[string]*ExcIEEEAC3A
    ExcIEEEAC4As map[string]*ExcIEEEAC4A
    ExcIEEEAC5As map[string]*ExcIEEEAC5A
    ExcIEEEAC6As map[string]*ExcIEEEAC6A
    ExcIEEEAC7Bs map[string]*ExcIEEEAC7B
    ExcIEEEAC8Bs map[string]*ExcIEEEAC8B
    ExcIEEEDC1As map[string]*ExcIEEEDC1A
    ExcIEEEDC2As map[string]*ExcIEEEDC2A
    ExcIEEEDC3As map[string]*ExcIEEEDC3A
    ExcIEEEDC4Bs map[string]*ExcIEEEDC4B
    ExcIEEEST1As map[string]*ExcIEEEST1A
    ExcIEEEST2As map[string]*ExcIEEEST2A
    ExcIEEEST3As map[string]*ExcIEEEST3A
    ExcIEEEST4Bs map[string]*ExcIEEEST4B
    ExcIEEEST5Bs map[string]*ExcIEEEST5B
    ExcIEEEST6Bs map[string]*ExcIEEEST6B
    ExcIEEEST7Bs map[string]*ExcIEEEST7B
    ExcNIs map[string]*ExcNI
    ExcOEX3Ts map[string]*ExcOEX3T
    ExcPICs map[string]*ExcPIC
    ExcREXSs map[string]*ExcREXS
    ExcRQBs map[string]*ExcRQB
    ExcSCRXs map[string]*ExcSCRX
    ExcSEXSs map[string]*ExcSEXS
    ExcSKs map[string]*ExcSK
    ExcST1As map[string]*ExcST1A
    ExcST2As map[string]*ExcST2A
    ExcST3As map[string]*ExcST3A
    ExcST4Bs map[string]*ExcST4B
    ExcST6Bs map[string]*ExcST6B
    ExcST7Bs map[string]*ExcST7B
    ExcitationSystemUserDefineds map[string]*ExcitationSystemUserDefined
    ExternalNetworkInjections map[string]*ExternalNetworkInjection
    FaultIndicators map[string]*FaultIndicator
    FossilFuels map[string]*FossilFuel
    FullModels map[string]*FullModel
    Fuses map[string]*Fuse
    GenICompensationForGenJs map[string]*GenICompensationForGenJ
    GeneratingUnits map[string]*GeneratingUnit
    GeographicalRegions map[string]*GeographicalRegion
    GovCT1s map[string]*GovCT1
    GovCT2s map[string]*GovCT2
    GovGASTs map[string]*GovGAST
    GovGAST1s map[string]*GovGAST1
    GovGAST2s map[string]*GovGAST2
    GovGAST3s map[string]*GovGAST3
    GovGAST4s map[string]*GovGAST4
    GovGASTWDs map[string]*GovGASTWD
    GovHydro1s map[string]*GovHydro1
    GovHydro2s map[string]*GovHydro2
    GovHydro3s map[string]*GovHydro3
    GovHydro4s map[string]*GovHydro4
    GovHydroDDs map[string]*GovHydroDD
    GovHydroFranciss map[string]*GovHydroFrancis
    GovHydroIEEE0s map[string]*GovHydroIEEE0
    GovHydroIEEE2s map[string]*GovHydroIEEE2
    GovHydroPIDs map[string]*GovHydroPID
    GovHydroPID2s map[string]*GovHydroPID2
    GovHydroPeltons map[string]*GovHydroPelton
    GovHydroRs map[string]*GovHydroR
    GovHydroWEHs map[string]*GovHydroWEH
    GovHydroWPIDs map[string]*GovHydroWPID
    GovSteam0s map[string]*GovSteam0
    GovSteam1s map[string]*GovSteam1
    GovSteam2s map[string]*GovSteam2
    GovSteamBBs map[string]*GovSteamBB
    GovSteamCCs map[string]*GovSteamCC
    GovSteamEUs map[string]*GovSteamEU
    GovSteamFV2s map[string]*GovSteamFV2
    GovSteamFV3s map[string]*GovSteamFV3
    GovSteamFV4s map[string]*GovSteamFV4
    GovSteamIEEE1s map[string]*GovSteamIEEE1
    GovSteamSGOs map[string]*GovSteamSGO
    GrossToNetActivePowerCurves map[string]*GrossToNetActivePowerCurve
    Grounds map[string]*Ground
    GroundDisconnectors map[string]*GroundDisconnector
    GroundingImpedances map[string]*GroundingImpedance
    HydroGeneratingUnits map[string]*HydroGeneratingUnit
    HydroPowerPlants map[string]*HydroPowerPlant
    HydroPumps map[string]*HydroPump
    Jumpers map[string]*Jumper
    Junctions map[string]*Junction
    Lines map[string]*Line
    LinearShuntCompensators map[string]*LinearShuntCompensator
    LoadAggregates map[string]*LoadAggregate
    LoadAreas map[string]*LoadArea
    LoadBreakSwitchs map[string]*LoadBreakSwitch
    LoadComposites map[string]*LoadComposite
    LoadGenericNonLinears map[string]*LoadGenericNonLinear
    LoadMotors map[string]*LoadMotor
    LoadResponseCharacteristics map[string]*LoadResponseCharacteristic
    LoadStatics map[string]*LoadStatic
    LoadUserDefineds map[string]*LoadUserDefined
    Locations map[string]*Location
    MeasurementValueQualitys map[string]*MeasurementValueQuality
    MeasurementValueSources map[string]*MeasurementValueSource
    MechLoad1s map[string]*MechLoad1
    MechanicalLoadUserDefineds map[string]*MechanicalLoadUserDefined
    MutualCouplings map[string]*MutualCoupling
    NonConformLoads map[string]*NonConformLoad
    NonConformLoadGroups map[string]*NonConformLoadGroup
    NonConformLoadSchedules map[string]*NonConformLoadSchedule
    NonlinearShuntCompensators map[string]*NonlinearShuntCompensator
    NonlinearShuntCompensatorPoints map[string]*NonlinearShuntCompensatorPoint
    NuclearGeneratingUnits map[string]*NuclearGeneratingUnit
    OperationalLimitSets map[string]*OperationalLimitSet
    OperationalLimitTypes map[string]*OperationalLimitType
    OverexcLim2s map[string]*OverexcLim2
    OverexcLimIEEEs map[string]*OverexcLimIEEE
    OverexcLimX1s map[string]*OverexcLimX1
    OverexcLimX2s map[string]*OverexcLimX2
    OverexcitationLimiterUserDefineds map[string]*OverexcitationLimiterUserDefined
    PFVArControllerType1UserDefineds map[string]*PFVArControllerType1UserDefined
    PFVArControllerType2UserDefineds map[string]*PFVArControllerType2UserDefined
    PFVArType1IEEEPFControllers map[string]*PFVArType1IEEEPFController
    PFVArType1IEEEVArControllers map[string]*PFVArType1IEEEVArController
    PFVArType2Common1s map[string]*PFVArType2Common1
    PFVArType2IEEEPFControllers map[string]*PFVArType2IEEEPFController
    PFVArType2IEEEVArControllers map[string]*PFVArType2IEEEVArController
    PetersenCoils map[string]*PetersenCoil
    PhaseTapChangerAsymmetricals map[string]*PhaseTapChangerAsymmetrical
    PhaseTapChangerLinears map[string]*PhaseTapChangerLinear
    PhaseTapChangerSymmetricals map[string]*PhaseTapChangerSymmetrical
    PhaseTapChangerTables map[string]*PhaseTapChangerTable
    PhaseTapChangerTablePoints map[string]*PhaseTapChangerTablePoint
    PhaseTapChangerTabulars map[string]*PhaseTapChangerTabular
    PhotoVoltaicUnits map[string]*PhotoVoltaicUnit
    PositionPoints map[string]*PositionPoint
    PostLineSensors map[string]*PostLineSensor
    PotentialTransformers map[string]*PotentialTransformer
    PowerElectronicsConnections map[string]*PowerElectronicsConnection
    PowerElectronicsWindUnits map[string]*PowerElectronicsWindUnit
    PowerSystemStabilizerUserDefineds map[string]*PowerSystemStabilizerUserDefined
    PowerTransformers map[string]*PowerTransformer
    PowerTransformerEnds map[string]*PowerTransformerEnd
    ProprietaryParameterDynamicss map[string]*ProprietaryParameterDynamics
    Pss1s map[string]*Pss1
    Pss1As map[string]*Pss1A
    Pss2Bs map[string]*Pss2B
    Pss2STs map[string]*Pss2ST
    Pss5s map[string]*Pss5
    PssELIN2s map[string]*PssELIN2
    PssIEEE1As map[string]*PssIEEE1A
    PssIEEE2Bs map[string]*PssIEEE2B
    PssIEEE3Bs map[string]*PssIEEE3B
    PssIEEE4Bs map[string]*PssIEEE4B
    PssPTIST1s map[string]*PssPTIST1
    PssPTIST3s map[string]*PssPTIST3
    PssRQBs map[string]*PssRQB
    PssSB4s map[string]*PssSB4
    PssSHs map[string]*PssSH
    PssSKs map[string]*PssSK
    PssSTAB2As map[string]*PssSTAB2A
    PssWECCs map[string]*PssWECC
    RaiseLowerCommands map[string]*RaiseLowerCommand
    RatioTapChangers map[string]*RatioTapChanger
    RatioTapChangerTables map[string]*RatioTapChangerTable
    RatioTapChangerTablePoints map[string]*RatioTapChangerTablePoint
    ReactiveCapabilityCurves map[string]*ReactiveCapabilityCurve
    RegularTimePoints map[string]*RegularTimePoint
    RegulatingControls map[string]*RegulatingControl
    RegulationSchedules map[string]*RegulationSchedule
    RemoteInputSignals map[string]*RemoteInputSignal
    ReportingGroups map[string]*ReportingGroup
    SVCUserDefineds map[string]*SVCUserDefined
    Seasons map[string]*Season
    SeriesCompensators map[string]*SeriesCompensator
    ServiceLocations map[string]*ServiceLocation
    SetPoints map[string]*SetPoint
    SolarGeneratingUnits map[string]*SolarGeneratingUnit
    SolarPowerPlants map[string]*SolarPowerPlant
    StaticVarCompensators map[string]*StaticVarCompensator
    StationSupplys map[string]*StationSupply
    StringMeasurements map[string]*StringMeasurement
    StringMeasurementValues map[string]*StringMeasurementValue
    SubGeographicalRegions map[string]*SubGeographicalRegion
    SubLoadAreas map[string]*SubLoadArea
    Substations map[string]*Substation
    SurgeArresters map[string]*SurgeArrester
    SvInjections map[string]*SvInjection
    SvPowerFlows map[string]*SvPowerFlow
    SvShuntCompensatorSectionss map[string]*SvShuntCompensatorSections
    SvStatuss map[string]*SvStatus
    SvSwitchs map[string]*SvSwitch
    SvTapSteps map[string]*SvTapStep
    SvVoltages map[string]*SvVoltage
    Switchs map[string]*Switch
    SwitchSchedules map[string]*SwitchSchedule
    SynchronousMachines map[string]*SynchronousMachine
    SynchronousMachineEquivalentCircuits map[string]*SynchronousMachineEquivalentCircuit
    SynchronousMachineSimplifieds map[string]*SynchronousMachineSimplified
    SynchronousMachineTimeConstantReactances map[string]*SynchronousMachineTimeConstantReactance
    SynchronousMachineUserDefineds map[string]*SynchronousMachineUserDefined
    TapChangerControls map[string]*TapChangerControl
    TapSchedules map[string]*TapSchedule
    Terminals map[string]*Terminal
    TextDiagramObjects map[string]*TextDiagramObject
    ThermalGeneratingUnits map[string]*ThermalGeneratingUnit
    TieFlows map[string]*TieFlow
    TopologicalIslands map[string]*TopologicalIsland
    TopologicalNodes map[string]*TopologicalNode
    TurbLCFB1s map[string]*TurbLCFB1
    TurbineGovernorUserDefineds map[string]*TurbineGovernorUserDefined
    TurbineLoadControllerUserDefineds map[string]*TurbineLoadControllerUserDefined
    UnderexcLim2Simplifieds map[string]*UnderexcLim2Simplified
    UnderexcLimIEEE1s map[string]*UnderexcLimIEEE1
    UnderexcLimIEEE2s map[string]*UnderexcLimIEEE2
    UnderexcLimX1s map[string]*UnderexcLimX1
    UnderexcLimX2s map[string]*UnderexcLimX2
    UnderexcitationLimiterUserDefineds map[string]*UnderexcitationLimiterUserDefined
    VAdjIEEEs map[string]*VAdjIEEE
    VCompIEEEType1s map[string]*VCompIEEEType1
    VCompIEEEType2s map[string]*VCompIEEEType2
    VSCUserDefineds map[string]*VSCUserDefined
    ValueAliasSets map[string]*ValueAliasSet
    ValueToAliass map[string]*ValueToAlias
    VisibilityLayers map[string]*VisibilityLayer
    VoltageAdjusterUserDefineds map[string]*VoltageAdjusterUserDefined
    VoltageCompensatorUserDefineds map[string]*VoltageCompensatorUserDefined
    VoltageLevels map[string]*VoltageLevel
    VoltageLimits map[string]*VoltageLimit
    VsCapabilityCurves map[string]*VsCapabilityCurve
    VsConverters map[string]*VsConverter
    WaveTraps map[string]*WaveTrap
    WindAeroConstIECs map[string]*WindAeroConstIEC
    WindAeroOneDimIECs map[string]*WindAeroOneDimIEC
    WindAeroTwoDimIECs map[string]*WindAeroTwoDimIEC
    WindContCurrLimIECs map[string]*WindContCurrLimIEC
    WindContPType3IECs map[string]*WindContPType3IEC
    WindContPType4aIECs map[string]*WindContPType4aIEC
    WindContPType4bIECs map[string]*WindContPType4bIEC
    WindContPitchAngleIECs map[string]*WindContPitchAngleIEC
    WindContQIECs map[string]*WindContQIEC
    WindContQLimIECs map[string]*WindContQLimIEC
    WindContQPQULimIECs map[string]*WindContQPQULimIEC
    WindContRotorRIECs map[string]*WindContRotorRIEC
    WindDynamicsLookupTables map[string]*WindDynamicsLookupTable
    WindGenTurbineType1aIECs map[string]*WindGenTurbineType1aIEC
    WindGenTurbineType1bIECs map[string]*WindGenTurbineType1bIEC
    WindGenTurbineType2IECs map[string]*WindGenTurbineType2IEC
    WindGenType3aIECs map[string]*WindGenType3aIEC
    WindGenType3bIECs map[string]*WindGenType3bIEC
    WindGenType4IECs map[string]*WindGenType4IEC
    WindGeneratingUnits map[string]*WindGeneratingUnit
    WindMechIECs map[string]*WindMechIEC
    WindPitchContPowerIECs map[string]*WindPitchContPowerIEC
    WindPlantFreqPcontrolIECs map[string]*WindPlantFreqPcontrolIEC
    WindPlantIECs map[string]*WindPlantIEC
    WindPlantReactiveControlIECs map[string]*WindPlantReactiveControlIEC
    WindPlantUserDefineds map[string]*WindPlantUserDefined
    WindPowerPlants map[string]*WindPowerPlant
    WindProtectionIECs map[string]*WindProtectionIEC
    WindRefFrameRotIECs map[string]*WindRefFrameRotIEC
    WindTurbineType3IECs map[string]*WindTurbineType3IEC
    WindTurbineType4aIECs map[string]*WindTurbineType4aIEC
    WindTurbineType4bIECs map[string]*WindTurbineType4bIEC
    WindType1or2UserDefineds map[string]*WindType1or2UserDefined
    WindType3or4UserDefineds map[string]*WindType3or4UserDefined
}

func NewCIMElementList() *CIMElementList {
	return &CIMElementList{
		Elements: make(map[string]interface{}),
		ACDCConverterDCTerminals: make(map[string]*ACDCConverterDCTerminal),
		ACLineSegments: make(map[string]*ACLineSegment),
		Accumulators: make(map[string]*Accumulator),
		AccumulatorLimits: make(map[string]*AccumulatorLimit),
		AccumulatorLimitSets: make(map[string]*AccumulatorLimitSet),
		AccumulatorResets: make(map[string]*AccumulatorReset),
		AccumulatorValues: make(map[string]*AccumulatorValue),
		ActivePowerLimits: make(map[string]*ActivePowerLimit),
		Analogs: make(map[string]*Analog),
		AnalogLimits: make(map[string]*AnalogLimit),
		AnalogLimitSets: make(map[string]*AnalogLimitSet),
		AnalogValues: make(map[string]*AnalogValue),
		ApparentPowerLimits: make(map[string]*ApparentPowerLimit),
		AsynchronousMachines: make(map[string]*AsynchronousMachine),
		AsynchronousMachineEquivalentCircuits: make(map[string]*AsynchronousMachineEquivalentCircuit),
		AsynchronousMachineTimeConstantReactances: make(map[string]*AsynchronousMachineTimeConstantReactance),
		AsynchronousMachineUserDefineds: make(map[string]*AsynchronousMachineUserDefined),
		BaseVoltages: make(map[string]*BaseVoltage),
		BatteryUnits: make(map[string]*BatteryUnit),
		Bays: make(map[string]*Bay),
		BoundaryPoints: make(map[string]*BoundaryPoint),
		Breakers: make(map[string]*Breaker),
		BusNameMarkers: make(map[string]*BusNameMarker),
		BusbarSections: make(map[string]*BusbarSection),
		CAESPlants: make(map[string]*CAESPlant),
		CSCUserDefineds: make(map[string]*CSCUserDefined),
		Clamps: make(map[string]*Clamp),
		CogenerationPlants: make(map[string]*CogenerationPlant),
		CombinedCyclePlants: make(map[string]*CombinedCyclePlant),
		Commands: make(map[string]*Command),
		ConformLoads: make(map[string]*ConformLoad),
		ConformLoadGroups: make(map[string]*ConformLoadGroup),
		ConformLoadSchedules: make(map[string]*ConformLoadSchedule),
		ConnectivityNodes: make(map[string]*ConnectivityNode),
		ControlAreas: make(map[string]*ControlArea),
		ControlAreaGeneratingUnits: make(map[string]*ControlAreaGeneratingUnit),
		CoordinateSystems: make(map[string]*CoordinateSystem),
		CsConverters: make(map[string]*CsConverter),
		CurrentLimits: make(map[string]*CurrentLimit),
		CurrentTransformers: make(map[string]*CurrentTransformer),
		CurveDatas: make(map[string]*CurveData),
		Cuts: make(map[string]*Cut),
		DCBreakers: make(map[string]*DCBreaker),
		DCBusbars: make(map[string]*DCBusbar),
		DCChoppers: make(map[string]*DCChopper),
		DCConverterUnits: make(map[string]*DCConverterUnit),
		DCDisconnectors: make(map[string]*DCDisconnector),
		DCGrounds: make(map[string]*DCGround),
		DCLines: make(map[string]*DCLine),
		DCLineSegments: make(map[string]*DCLineSegment),
		DCNodes: make(map[string]*DCNode),
		DCSeriesDevices: make(map[string]*DCSeriesDevice),
		DCShunts: make(map[string]*DCShunt),
		DCSwitchs: make(map[string]*DCSwitch),
		DCTerminals: make(map[string]*DCTerminal),
		DCTopologicalIslands: make(map[string]*DCTopologicalIsland),
		DCTopologicalNodes: make(map[string]*DCTopologicalNode),
		DayTypes: make(map[string]*DayType),
		Diagrams: make(map[string]*Diagram),
		DiagramObjects: make(map[string]*DiagramObject),
		DiagramObjectGluePoints: make(map[string]*DiagramObjectGluePoint),
		DiagramObjectPoints: make(map[string]*DiagramObjectPoint),
		DiagramObjectStyles: make(map[string]*DiagramObjectStyle),
		DiagramStyles: make(map[string]*DiagramStyle),
		DifferenceModels: make(map[string]*DifferenceModel),
		DiscExcContIEEEDEC1As: make(map[string]*DiscExcContIEEEDEC1A),
		DiscExcContIEEEDEC2As: make(map[string]*DiscExcContIEEEDEC2A),
		DiscExcContIEEEDEC3As: make(map[string]*DiscExcContIEEEDEC3A),
		DisconnectingCircuitBreakers: make(map[string]*DisconnectingCircuitBreaker),
		Disconnectors: make(map[string]*Disconnector),
		DiscontinuousExcitationControlUserDefineds: make(map[string]*DiscontinuousExcitationControlUserDefined),
		Discretes: make(map[string]*Discrete),
		DiscreteValues: make(map[string]*DiscreteValue),
		EnergyConsumers: make(map[string]*EnergyConsumer),
		EnergySchedulingTypes: make(map[string]*EnergySchedulingType),
		EnergySources: make(map[string]*EnergySource),
		Equipments: make(map[string]*Equipment),
		EquivalentBranchs: make(map[string]*EquivalentBranch),
		EquivalentInjections: make(map[string]*EquivalentInjection),
		EquivalentNetworks: make(map[string]*EquivalentNetwork),
		EquivalentShunts: make(map[string]*EquivalentShunt),
		ExcAC1As: make(map[string]*ExcAC1A),
		ExcAC2As: make(map[string]*ExcAC2A),
		ExcAC3As: make(map[string]*ExcAC3A),
		ExcAC4As: make(map[string]*ExcAC4A),
		ExcAC5As: make(map[string]*ExcAC5A),
		ExcAC6As: make(map[string]*ExcAC6A),
		ExcAC8Bs: make(map[string]*ExcAC8B),
		ExcANSs: make(map[string]*ExcANS),
		ExcAVR1s: make(map[string]*ExcAVR1),
		ExcAVR2s: make(map[string]*ExcAVR2),
		ExcAVR3s: make(map[string]*ExcAVR3),
		ExcAVR4s: make(map[string]*ExcAVR4),
		ExcAVR5s: make(map[string]*ExcAVR5),
		ExcAVR7s: make(map[string]*ExcAVR7),
		ExcBBCs: make(map[string]*ExcBBC),
		ExcCZs: make(map[string]*ExcCZ),
		ExcDC1As: make(map[string]*ExcDC1A),
		ExcDC2As: make(map[string]*ExcDC2A),
		ExcDC3As: make(map[string]*ExcDC3A),
		ExcDC3A1s: make(map[string]*ExcDC3A1),
		ExcELIN1s: make(map[string]*ExcELIN1),
		ExcELIN2s: make(map[string]*ExcELIN2),
		ExcHUs: make(map[string]*ExcHU),
		ExcIEEEAC1As: make(map[string]*ExcIEEEAC1A),
		ExcIEEEAC2As: make(map[string]*ExcIEEEAC2A),
		ExcIEEEAC3As: make(map[string]*ExcIEEEAC3A),
		ExcIEEEAC4As: make(map[string]*ExcIEEEAC4A),
		ExcIEEEAC5As: make(map[string]*ExcIEEEAC5A),
		ExcIEEEAC6As: make(map[string]*ExcIEEEAC6A),
		ExcIEEEAC7Bs: make(map[string]*ExcIEEEAC7B),
		ExcIEEEAC8Bs: make(map[string]*ExcIEEEAC8B),
		ExcIEEEDC1As: make(map[string]*ExcIEEEDC1A),
		ExcIEEEDC2As: make(map[string]*ExcIEEEDC2A),
		ExcIEEEDC3As: make(map[string]*ExcIEEEDC3A),
		ExcIEEEDC4Bs: make(map[string]*ExcIEEEDC4B),
		ExcIEEEST1As: make(map[string]*ExcIEEEST1A),
		ExcIEEEST2As: make(map[string]*ExcIEEEST2A),
		ExcIEEEST3As: make(map[string]*ExcIEEEST3A),
		ExcIEEEST4Bs: make(map[string]*ExcIEEEST4B),
		ExcIEEEST5Bs: make(map[string]*ExcIEEEST5B),
		ExcIEEEST6Bs: make(map[string]*ExcIEEEST6B),
		ExcIEEEST7Bs: make(map[string]*ExcIEEEST7B),
		ExcNIs: make(map[string]*ExcNI),
		ExcOEX3Ts: make(map[string]*ExcOEX3T),
		ExcPICs: make(map[string]*ExcPIC),
		ExcREXSs: make(map[string]*ExcREXS),
		ExcRQBs: make(map[string]*ExcRQB),
		ExcSCRXs: make(map[string]*ExcSCRX),
		ExcSEXSs: make(map[string]*ExcSEXS),
		ExcSKs: make(map[string]*ExcSK),
		ExcST1As: make(map[string]*ExcST1A),
		ExcST2As: make(map[string]*ExcST2A),
		ExcST3As: make(map[string]*ExcST3A),
		ExcST4Bs: make(map[string]*ExcST4B),
		ExcST6Bs: make(map[string]*ExcST6B),
		ExcST7Bs: make(map[string]*ExcST7B),
		ExcitationSystemUserDefineds: make(map[string]*ExcitationSystemUserDefined),
		ExternalNetworkInjections: make(map[string]*ExternalNetworkInjection),
		FaultIndicators: make(map[string]*FaultIndicator),
		FossilFuels: make(map[string]*FossilFuel),
		FullModels: make(map[string]*FullModel),
		Fuses: make(map[string]*Fuse),
		GenICompensationForGenJs: make(map[string]*GenICompensationForGenJ),
		GeneratingUnits: make(map[string]*GeneratingUnit),
		GeographicalRegions: make(map[string]*GeographicalRegion),
		GovCT1s: make(map[string]*GovCT1),
		GovCT2s: make(map[string]*GovCT2),
		GovGASTs: make(map[string]*GovGAST),
		GovGAST1s: make(map[string]*GovGAST1),
		GovGAST2s: make(map[string]*GovGAST2),
		GovGAST3s: make(map[string]*GovGAST3),
		GovGAST4s: make(map[string]*GovGAST4),
		GovGASTWDs: make(map[string]*GovGASTWD),
		GovHydro1s: make(map[string]*GovHydro1),
		GovHydro2s: make(map[string]*GovHydro2),
		GovHydro3s: make(map[string]*GovHydro3),
		GovHydro4s: make(map[string]*GovHydro4),
		GovHydroDDs: make(map[string]*GovHydroDD),
		GovHydroFranciss: make(map[string]*GovHydroFrancis),
		GovHydroIEEE0s: make(map[string]*GovHydroIEEE0),
		GovHydroIEEE2s: make(map[string]*GovHydroIEEE2),
		GovHydroPIDs: make(map[string]*GovHydroPID),
		GovHydroPID2s: make(map[string]*GovHydroPID2),
		GovHydroPeltons: make(map[string]*GovHydroPelton),
		GovHydroRs: make(map[string]*GovHydroR),
		GovHydroWEHs: make(map[string]*GovHydroWEH),
		GovHydroWPIDs: make(map[string]*GovHydroWPID),
		GovSteam0s: make(map[string]*GovSteam0),
		GovSteam1s: make(map[string]*GovSteam1),
		GovSteam2s: make(map[string]*GovSteam2),
		GovSteamBBs: make(map[string]*GovSteamBB),
		GovSteamCCs: make(map[string]*GovSteamCC),
		GovSteamEUs: make(map[string]*GovSteamEU),
		GovSteamFV2s: make(map[string]*GovSteamFV2),
		GovSteamFV3s: make(map[string]*GovSteamFV3),
		GovSteamFV4s: make(map[string]*GovSteamFV4),
		GovSteamIEEE1s: make(map[string]*GovSteamIEEE1),
		GovSteamSGOs: make(map[string]*GovSteamSGO),
		GrossToNetActivePowerCurves: make(map[string]*GrossToNetActivePowerCurve),
		Grounds: make(map[string]*Ground),
		GroundDisconnectors: make(map[string]*GroundDisconnector),
		GroundingImpedances: make(map[string]*GroundingImpedance),
		HydroGeneratingUnits: make(map[string]*HydroGeneratingUnit),
		HydroPowerPlants: make(map[string]*HydroPowerPlant),
		HydroPumps: make(map[string]*HydroPump),
		Jumpers: make(map[string]*Jumper),
		Junctions: make(map[string]*Junction),
		Lines: make(map[string]*Line),
		LinearShuntCompensators: make(map[string]*LinearShuntCompensator),
		LoadAggregates: make(map[string]*LoadAggregate),
		LoadAreas: make(map[string]*LoadArea),
		LoadBreakSwitchs: make(map[string]*LoadBreakSwitch),
		LoadComposites: make(map[string]*LoadComposite),
		LoadGenericNonLinears: make(map[string]*LoadGenericNonLinear),
		LoadMotors: make(map[string]*LoadMotor),
		LoadResponseCharacteristics: make(map[string]*LoadResponseCharacteristic),
		LoadStatics: make(map[string]*LoadStatic),
		LoadUserDefineds: make(map[string]*LoadUserDefined),
		Locations: make(map[string]*Location),
		MeasurementValueQualitys: make(map[string]*MeasurementValueQuality),
		MeasurementValueSources: make(map[string]*MeasurementValueSource),
		MechLoad1s: make(map[string]*MechLoad1),
		MechanicalLoadUserDefineds: make(map[string]*MechanicalLoadUserDefined),
		MutualCouplings: make(map[string]*MutualCoupling),
		NonConformLoads: make(map[string]*NonConformLoad),
		NonConformLoadGroups: make(map[string]*NonConformLoadGroup),
		NonConformLoadSchedules: make(map[string]*NonConformLoadSchedule),
		NonlinearShuntCompensators: make(map[string]*NonlinearShuntCompensator),
		NonlinearShuntCompensatorPoints: make(map[string]*NonlinearShuntCompensatorPoint),
		NuclearGeneratingUnits: make(map[string]*NuclearGeneratingUnit),
		OperationalLimitSets: make(map[string]*OperationalLimitSet),
		OperationalLimitTypes: make(map[string]*OperationalLimitType),
		OverexcLim2s: make(map[string]*OverexcLim2),
		OverexcLimIEEEs: make(map[string]*OverexcLimIEEE),
		OverexcLimX1s: make(map[string]*OverexcLimX1),
		OverexcLimX2s: make(map[string]*OverexcLimX2),
		OverexcitationLimiterUserDefineds: make(map[string]*OverexcitationLimiterUserDefined),
		PFVArControllerType1UserDefineds: make(map[string]*PFVArControllerType1UserDefined),
		PFVArControllerType2UserDefineds: make(map[string]*PFVArControllerType2UserDefined),
		PFVArType1IEEEPFControllers: make(map[string]*PFVArType1IEEEPFController),
		PFVArType1IEEEVArControllers: make(map[string]*PFVArType1IEEEVArController),
		PFVArType2Common1s: make(map[string]*PFVArType2Common1),
		PFVArType2IEEEPFControllers: make(map[string]*PFVArType2IEEEPFController),
		PFVArType2IEEEVArControllers: make(map[string]*PFVArType2IEEEVArController),
		PetersenCoils: make(map[string]*PetersenCoil),
		PhaseTapChangerAsymmetricals: make(map[string]*PhaseTapChangerAsymmetrical),
		PhaseTapChangerLinears: make(map[string]*PhaseTapChangerLinear),
		PhaseTapChangerSymmetricals: make(map[string]*PhaseTapChangerSymmetrical),
		PhaseTapChangerTables: make(map[string]*PhaseTapChangerTable),
		PhaseTapChangerTablePoints: make(map[string]*PhaseTapChangerTablePoint),
		PhaseTapChangerTabulars: make(map[string]*PhaseTapChangerTabular),
		PhotoVoltaicUnits: make(map[string]*PhotoVoltaicUnit),
		PositionPoints: make(map[string]*PositionPoint),
		PostLineSensors: make(map[string]*PostLineSensor),
		PotentialTransformers: make(map[string]*PotentialTransformer),
		PowerElectronicsConnections: make(map[string]*PowerElectronicsConnection),
		PowerElectronicsWindUnits: make(map[string]*PowerElectronicsWindUnit),
		PowerSystemStabilizerUserDefineds: make(map[string]*PowerSystemStabilizerUserDefined),
		PowerTransformers: make(map[string]*PowerTransformer),
		PowerTransformerEnds: make(map[string]*PowerTransformerEnd),
		ProprietaryParameterDynamicss: make(map[string]*ProprietaryParameterDynamics),
		Pss1s: make(map[string]*Pss1),
		Pss1As: make(map[string]*Pss1A),
		Pss2Bs: make(map[string]*Pss2B),
		Pss2STs: make(map[string]*Pss2ST),
		Pss5s: make(map[string]*Pss5),
		PssELIN2s: make(map[string]*PssELIN2),
		PssIEEE1As: make(map[string]*PssIEEE1A),
		PssIEEE2Bs: make(map[string]*PssIEEE2B),
		PssIEEE3Bs: make(map[string]*PssIEEE3B),
		PssIEEE4Bs: make(map[string]*PssIEEE4B),
		PssPTIST1s: make(map[string]*PssPTIST1),
		PssPTIST3s: make(map[string]*PssPTIST3),
		PssRQBs: make(map[string]*PssRQB),
		PssSB4s: make(map[string]*PssSB4),
		PssSHs: make(map[string]*PssSH),
		PssSKs: make(map[string]*PssSK),
		PssSTAB2As: make(map[string]*PssSTAB2A),
		PssWECCs: make(map[string]*PssWECC),
		RaiseLowerCommands: make(map[string]*RaiseLowerCommand),
		RatioTapChangers: make(map[string]*RatioTapChanger),
		RatioTapChangerTables: make(map[string]*RatioTapChangerTable),
		RatioTapChangerTablePoints: make(map[string]*RatioTapChangerTablePoint),
		ReactiveCapabilityCurves: make(map[string]*ReactiveCapabilityCurve),
		RegularTimePoints: make(map[string]*RegularTimePoint),
		RegulatingControls: make(map[string]*RegulatingControl),
		RegulationSchedules: make(map[string]*RegulationSchedule),
		RemoteInputSignals: make(map[string]*RemoteInputSignal),
		ReportingGroups: make(map[string]*ReportingGroup),
		SVCUserDefineds: make(map[string]*SVCUserDefined),
		Seasons: make(map[string]*Season),
		SeriesCompensators: make(map[string]*SeriesCompensator),
		ServiceLocations: make(map[string]*ServiceLocation),
		SetPoints: make(map[string]*SetPoint),
		SolarGeneratingUnits: make(map[string]*SolarGeneratingUnit),
		SolarPowerPlants: make(map[string]*SolarPowerPlant),
		StaticVarCompensators: make(map[string]*StaticVarCompensator),
		StationSupplys: make(map[string]*StationSupply),
		StringMeasurements: make(map[string]*StringMeasurement),
		StringMeasurementValues: make(map[string]*StringMeasurementValue),
		SubGeographicalRegions: make(map[string]*SubGeographicalRegion),
		SubLoadAreas: make(map[string]*SubLoadArea),
		Substations: make(map[string]*Substation),
		SurgeArresters: make(map[string]*SurgeArrester),
		SvInjections: make(map[string]*SvInjection),
		SvPowerFlows: make(map[string]*SvPowerFlow),
		SvShuntCompensatorSectionss: make(map[string]*SvShuntCompensatorSections),
		SvStatuss: make(map[string]*SvStatus),
		SvSwitchs: make(map[string]*SvSwitch),
		SvTapSteps: make(map[string]*SvTapStep),
		SvVoltages: make(map[string]*SvVoltage),
		Switchs: make(map[string]*Switch),
		SwitchSchedules: make(map[string]*SwitchSchedule),
		SynchronousMachines: make(map[string]*SynchronousMachine),
		SynchronousMachineEquivalentCircuits: make(map[string]*SynchronousMachineEquivalentCircuit),
		SynchronousMachineSimplifieds: make(map[string]*SynchronousMachineSimplified),
		SynchronousMachineTimeConstantReactances: make(map[string]*SynchronousMachineTimeConstantReactance),
		SynchronousMachineUserDefineds: make(map[string]*SynchronousMachineUserDefined),
		TapChangerControls: make(map[string]*TapChangerControl),
		TapSchedules: make(map[string]*TapSchedule),
		Terminals: make(map[string]*Terminal),
		TextDiagramObjects: make(map[string]*TextDiagramObject),
		ThermalGeneratingUnits: make(map[string]*ThermalGeneratingUnit),
		TieFlows: make(map[string]*TieFlow),
		TopologicalIslands: make(map[string]*TopologicalIsland),
		TopologicalNodes: make(map[string]*TopologicalNode),
		TurbLCFB1s: make(map[string]*TurbLCFB1),
		TurbineGovernorUserDefineds: make(map[string]*TurbineGovernorUserDefined),
		TurbineLoadControllerUserDefineds: make(map[string]*TurbineLoadControllerUserDefined),
		UnderexcLim2Simplifieds: make(map[string]*UnderexcLim2Simplified),
		UnderexcLimIEEE1s: make(map[string]*UnderexcLimIEEE1),
		UnderexcLimIEEE2s: make(map[string]*UnderexcLimIEEE2),
		UnderexcLimX1s: make(map[string]*UnderexcLimX1),
		UnderexcLimX2s: make(map[string]*UnderexcLimX2),
		UnderexcitationLimiterUserDefineds: make(map[string]*UnderexcitationLimiterUserDefined),
		VAdjIEEEs: make(map[string]*VAdjIEEE),
		VCompIEEEType1s: make(map[string]*VCompIEEEType1),
		VCompIEEEType2s: make(map[string]*VCompIEEEType2),
		VSCUserDefineds: make(map[string]*VSCUserDefined),
		ValueAliasSets: make(map[string]*ValueAliasSet),
		ValueToAliass: make(map[string]*ValueToAlias),
		VisibilityLayers: make(map[string]*VisibilityLayer),
		VoltageAdjusterUserDefineds: make(map[string]*VoltageAdjusterUserDefined),
		VoltageCompensatorUserDefineds: make(map[string]*VoltageCompensatorUserDefined),
		VoltageLevels: make(map[string]*VoltageLevel),
		VoltageLimits: make(map[string]*VoltageLimit),
		VsCapabilityCurves: make(map[string]*VsCapabilityCurve),
		VsConverters: make(map[string]*VsConverter),
		WaveTraps: make(map[string]*WaveTrap),
		WindAeroConstIECs: make(map[string]*WindAeroConstIEC),
		WindAeroOneDimIECs: make(map[string]*WindAeroOneDimIEC),
		WindAeroTwoDimIECs: make(map[string]*WindAeroTwoDimIEC),
		WindContCurrLimIECs: make(map[string]*WindContCurrLimIEC),
		WindContPType3IECs: make(map[string]*WindContPType3IEC),
		WindContPType4aIECs: make(map[string]*WindContPType4aIEC),
		WindContPType4bIECs: make(map[string]*WindContPType4bIEC),
		WindContPitchAngleIECs: make(map[string]*WindContPitchAngleIEC),
		WindContQIECs: make(map[string]*WindContQIEC),
		WindContQLimIECs: make(map[string]*WindContQLimIEC),
		WindContQPQULimIECs: make(map[string]*WindContQPQULimIEC),
		WindContRotorRIECs: make(map[string]*WindContRotorRIEC),
		WindDynamicsLookupTables: make(map[string]*WindDynamicsLookupTable),
		WindGenTurbineType1aIECs: make(map[string]*WindGenTurbineType1aIEC),
		WindGenTurbineType1bIECs: make(map[string]*WindGenTurbineType1bIEC),
		WindGenTurbineType2IECs: make(map[string]*WindGenTurbineType2IEC),
		WindGenType3aIECs: make(map[string]*WindGenType3aIEC),
		WindGenType3bIECs: make(map[string]*WindGenType3bIEC),
		WindGenType4IECs: make(map[string]*WindGenType4IEC),
		WindGeneratingUnits: make(map[string]*WindGeneratingUnit),
		WindMechIECs: make(map[string]*WindMechIEC),
		WindPitchContPowerIECs: make(map[string]*WindPitchContPowerIEC),
		WindPlantFreqPcontrolIECs: make(map[string]*WindPlantFreqPcontrolIEC),
		WindPlantIECs: make(map[string]*WindPlantIEC),
		WindPlantReactiveControlIECs: make(map[string]*WindPlantReactiveControlIEC),
		WindPlantUserDefineds: make(map[string]*WindPlantUserDefined),
		WindPowerPlants: make(map[string]*WindPowerPlant),
		WindProtectionIECs: make(map[string]*WindProtectionIEC),
		WindRefFrameRotIECs: make(map[string]*WindRefFrameRotIEC),
		WindTurbineType3IECs: make(map[string]*WindTurbineType3IEC),
		WindTurbineType4aIECs: make(map[string]*WindTurbineType4aIEC),
		WindTurbineType4bIECs: make(map[string]*WindTurbineType4bIEC),
		WindType1or2UserDefineds: make(map[string]*WindType1or2UserDefined),
		WindType3or4UserDefineds: make(map[string]*WindType3or4UserDefined),
	}
}

func (ds *CIMElementList) AddElement(element interface{}) {
    switch e := element.(type) {
	case *ACDCConverterDCTerminal:
		ds.ACDCConverterDCTerminals[e.Id] = e
		ds.Elements[e.Id] = e
	case *ACLineSegment:
		ds.ACLineSegments[e.Id] = e
		ds.Elements[e.Id] = e
	case *Accumulator:
		ds.Accumulators[e.Id] = e
		ds.Elements[e.Id] = e
	case *AccumulatorLimit:
		ds.AccumulatorLimits[e.Id] = e
		ds.Elements[e.Id] = e
	case *AccumulatorLimitSet:
		ds.AccumulatorLimitSets[e.Id] = e
		ds.Elements[e.Id] = e
	case *AccumulatorReset:
		ds.AccumulatorResets[e.Id] = e
		ds.Elements[e.Id] = e
	case *AccumulatorValue:
		ds.AccumulatorValues[e.Id] = e
		ds.Elements[e.Id] = e
	case *ActivePowerLimit:
		ds.ActivePowerLimits[e.Id] = e
		ds.Elements[e.Id] = e
	case *Analog:
		ds.Analogs[e.Id] = e
		ds.Elements[e.Id] = e
	case *AnalogLimit:
		ds.AnalogLimits[e.Id] = e
		ds.Elements[e.Id] = e
	case *AnalogLimitSet:
		ds.AnalogLimitSets[e.Id] = e
		ds.Elements[e.Id] = e
	case *AnalogValue:
		ds.AnalogValues[e.Id] = e
		ds.Elements[e.Id] = e
	case *ApparentPowerLimit:
		ds.ApparentPowerLimits[e.Id] = e
		ds.Elements[e.Id] = e
	case *AsynchronousMachine:
		ds.AsynchronousMachines[e.Id] = e
		ds.Elements[e.Id] = e
	case *AsynchronousMachineEquivalentCircuit:
		ds.AsynchronousMachineEquivalentCircuits[e.Id] = e
		ds.Elements[e.Id] = e
	case *AsynchronousMachineTimeConstantReactance:
		ds.AsynchronousMachineTimeConstantReactances[e.Id] = e
		ds.Elements[e.Id] = e
	case *AsynchronousMachineUserDefined:
		ds.AsynchronousMachineUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *BaseVoltage:
		ds.BaseVoltages[e.Id] = e
		ds.Elements[e.Id] = e
	case *BatteryUnit:
		ds.BatteryUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *Bay:
		ds.Bays[e.Id] = e
		ds.Elements[e.Id] = e
	case *BoundaryPoint:
		ds.BoundaryPoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *Breaker:
		ds.Breakers[e.Id] = e
		ds.Elements[e.Id] = e
	case *BusNameMarker:
		ds.BusNameMarkers[e.Id] = e
		ds.Elements[e.Id] = e
	case *BusbarSection:
		ds.BusbarSections[e.Id] = e
		ds.Elements[e.Id] = e
	case *CAESPlant:
		ds.CAESPlants[e.Id] = e
		ds.Elements[e.Id] = e
	case *CSCUserDefined:
		ds.CSCUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *Clamp:
		ds.Clamps[e.Id] = e
		ds.Elements[e.Id] = e
	case *CogenerationPlant:
		ds.CogenerationPlants[e.Id] = e
		ds.Elements[e.Id] = e
	case *CombinedCyclePlant:
		ds.CombinedCyclePlants[e.Id] = e
		ds.Elements[e.Id] = e
	case *Command:
		ds.Commands[e.Id] = e
		ds.Elements[e.Id] = e
	case *ConformLoad:
		ds.ConformLoads[e.Id] = e
		ds.Elements[e.Id] = e
	case *ConformLoadGroup:
		ds.ConformLoadGroups[e.Id] = e
		ds.Elements[e.Id] = e
	case *ConformLoadSchedule:
		ds.ConformLoadSchedules[e.Id] = e
		ds.Elements[e.Id] = e
	case *ConnectivityNode:
		ds.ConnectivityNodes[e.Id] = e
		ds.Elements[e.Id] = e
	case *ControlArea:
		ds.ControlAreas[e.Id] = e
		ds.Elements[e.Id] = e
	case *ControlAreaGeneratingUnit:
		ds.ControlAreaGeneratingUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *CoordinateSystem:
		ds.CoordinateSystems[e.Id] = e
		ds.Elements[e.Id] = e
	case *CsConverter:
		ds.CsConverters[e.Id] = e
		ds.Elements[e.Id] = e
	case *CurrentLimit:
		ds.CurrentLimits[e.Id] = e
		ds.Elements[e.Id] = e
	case *CurrentTransformer:
		ds.CurrentTransformers[e.Id] = e
		ds.Elements[e.Id] = e
	case *CurveData:
		ds.CurveDatas[e.Id] = e
		ds.Elements[e.Id] = e
	case *Cut:
		ds.Cuts[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCBreaker:
		ds.DCBreakers[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCBusbar:
		ds.DCBusbars[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCChopper:
		ds.DCChoppers[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCConverterUnit:
		ds.DCConverterUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCDisconnector:
		ds.DCDisconnectors[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCGround:
		ds.DCGrounds[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCLine:
		ds.DCLines[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCLineSegment:
		ds.DCLineSegments[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCNode:
		ds.DCNodes[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCSeriesDevice:
		ds.DCSeriesDevices[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCShunt:
		ds.DCShunts[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCSwitch:
		ds.DCSwitchs[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCTerminal:
		ds.DCTerminals[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCTopologicalIsland:
		ds.DCTopologicalIslands[e.Id] = e
		ds.Elements[e.Id] = e
	case *DCTopologicalNode:
		ds.DCTopologicalNodes[e.Id] = e
		ds.Elements[e.Id] = e
	case *DayType:
		ds.DayTypes[e.Id] = e
		ds.Elements[e.Id] = e
	case *Diagram:
		ds.Diagrams[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiagramObject:
		ds.DiagramObjects[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiagramObjectGluePoint:
		ds.DiagramObjectGluePoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiagramObjectPoint:
		ds.DiagramObjectPoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiagramObjectStyle:
		ds.DiagramObjectStyles[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiagramStyle:
		ds.DiagramStyles[e.Id] = e
		ds.Elements[e.Id] = e
	case *DifferenceModel:
		ds.DifferenceModels[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiscExcContIEEEDEC1A:
		ds.DiscExcContIEEEDEC1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiscExcContIEEEDEC2A:
		ds.DiscExcContIEEEDEC2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiscExcContIEEEDEC3A:
		ds.DiscExcContIEEEDEC3As[e.Id] = e
		ds.Elements[e.Id] = e
	case *DisconnectingCircuitBreaker:
		ds.DisconnectingCircuitBreakers[e.Id] = e
		ds.Elements[e.Id] = e
	case *Disconnector:
		ds.Disconnectors[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiscontinuousExcitationControlUserDefined:
		ds.DiscontinuousExcitationControlUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *Discrete:
		ds.Discretes[e.Id] = e
		ds.Elements[e.Id] = e
	case *DiscreteValue:
		ds.DiscreteValues[e.Id] = e
		ds.Elements[e.Id] = e
	case *EnergyConsumer:
		ds.EnergyConsumers[e.Id] = e
		ds.Elements[e.Id] = e
	case *EnergySchedulingType:
		ds.EnergySchedulingTypes[e.Id] = e
		ds.Elements[e.Id] = e
	case *EnergySource:
		ds.EnergySources[e.Id] = e
		ds.Elements[e.Id] = e
	case *Equipment:
		ds.Equipments[e.Id] = e
		ds.Elements[e.Id] = e
	case *EquivalentBranch:
		ds.EquivalentBranchs[e.Id] = e
		ds.Elements[e.Id] = e
	case *EquivalentInjection:
		ds.EquivalentInjections[e.Id] = e
		ds.Elements[e.Id] = e
	case *EquivalentNetwork:
		ds.EquivalentNetworks[e.Id] = e
		ds.Elements[e.Id] = e
	case *EquivalentShunt:
		ds.EquivalentShunts[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAC1A:
		ds.ExcAC1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAC2A:
		ds.ExcAC2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAC3A:
		ds.ExcAC3As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAC4A:
		ds.ExcAC4As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAC5A:
		ds.ExcAC5As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAC6A:
		ds.ExcAC6As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAC8B:
		ds.ExcAC8Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcANS:
		ds.ExcANSs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAVR1:
		ds.ExcAVR1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAVR2:
		ds.ExcAVR2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAVR3:
		ds.ExcAVR3s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAVR4:
		ds.ExcAVR4s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAVR5:
		ds.ExcAVR5s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcAVR7:
		ds.ExcAVR7s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcBBC:
		ds.ExcBBCs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcCZ:
		ds.ExcCZs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcDC1A:
		ds.ExcDC1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcDC2A:
		ds.ExcDC2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcDC3A:
		ds.ExcDC3As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcDC3A1:
		ds.ExcDC3A1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcELIN1:
		ds.ExcELIN1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcELIN2:
		ds.ExcELIN2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcHU:
		ds.ExcHUs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC1A:
		ds.ExcIEEEAC1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC2A:
		ds.ExcIEEEAC2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC3A:
		ds.ExcIEEEAC3As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC4A:
		ds.ExcIEEEAC4As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC5A:
		ds.ExcIEEEAC5As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC6A:
		ds.ExcIEEEAC6As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC7B:
		ds.ExcIEEEAC7Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEAC8B:
		ds.ExcIEEEAC8Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEDC1A:
		ds.ExcIEEEDC1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEDC2A:
		ds.ExcIEEEDC2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEDC3A:
		ds.ExcIEEEDC3As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEDC4B:
		ds.ExcIEEEDC4Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEST1A:
		ds.ExcIEEEST1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEST2A:
		ds.ExcIEEEST2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEST3A:
		ds.ExcIEEEST3As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEST4B:
		ds.ExcIEEEST4Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEST5B:
		ds.ExcIEEEST5Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEST6B:
		ds.ExcIEEEST6Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcIEEEST7B:
		ds.ExcIEEEST7Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcNI:
		ds.ExcNIs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcOEX3T:
		ds.ExcOEX3Ts[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcPIC:
		ds.ExcPICs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcREXS:
		ds.ExcREXSs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcRQB:
		ds.ExcRQBs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcSCRX:
		ds.ExcSCRXs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcSEXS:
		ds.ExcSEXSs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcSK:
		ds.ExcSKs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcST1A:
		ds.ExcST1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcST2A:
		ds.ExcST2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcST3A:
		ds.ExcST3As[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcST4B:
		ds.ExcST4Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcST6B:
		ds.ExcST6Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcST7B:
		ds.ExcST7Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExcitationSystemUserDefined:
		ds.ExcitationSystemUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *ExternalNetworkInjection:
		ds.ExternalNetworkInjections[e.Id] = e
		ds.Elements[e.Id] = e
	case *FaultIndicator:
		ds.FaultIndicators[e.Id] = e
		ds.Elements[e.Id] = e
	case *FossilFuel:
		ds.FossilFuels[e.Id] = e
		ds.Elements[e.Id] = e
	case *FullModel:
		ds.FullModels[e.Id] = e
		ds.Elements[e.Id] = e
	case *Fuse:
		ds.Fuses[e.Id] = e
		ds.Elements[e.Id] = e
	case *GenICompensationForGenJ:
		ds.GenICompensationForGenJs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GeneratingUnit:
		ds.GeneratingUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *GeographicalRegion:
		ds.GeographicalRegions[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovCT1:
		ds.GovCT1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovCT2:
		ds.GovCT2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovGAST:
		ds.GovGASTs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovGAST1:
		ds.GovGAST1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovGAST2:
		ds.GovGAST2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovGAST3:
		ds.GovGAST3s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovGAST4:
		ds.GovGAST4s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovGASTWD:
		ds.GovGASTWDs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydro1:
		ds.GovHydro1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydro2:
		ds.GovHydro2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydro3:
		ds.GovHydro3s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydro4:
		ds.GovHydro4s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroDD:
		ds.GovHydroDDs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroFrancis:
		ds.GovHydroFranciss[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroIEEE0:
		ds.GovHydroIEEE0s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroIEEE2:
		ds.GovHydroIEEE2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroPID:
		ds.GovHydroPIDs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroPID2:
		ds.GovHydroPID2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroPelton:
		ds.GovHydroPeltons[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroR:
		ds.GovHydroRs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroWEH:
		ds.GovHydroWEHs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovHydroWPID:
		ds.GovHydroWPIDs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteam0:
		ds.GovSteam0s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteam1:
		ds.GovSteam1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteam2:
		ds.GovSteam2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamBB:
		ds.GovSteamBBs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamCC:
		ds.GovSteamCCs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamEU:
		ds.GovSteamEUs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamFV2:
		ds.GovSteamFV2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamFV3:
		ds.GovSteamFV3s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamFV4:
		ds.GovSteamFV4s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamIEEE1:
		ds.GovSteamIEEE1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *GovSteamSGO:
		ds.GovSteamSGOs[e.Id] = e
		ds.Elements[e.Id] = e
	case *GrossToNetActivePowerCurve:
		ds.GrossToNetActivePowerCurves[e.Id] = e
		ds.Elements[e.Id] = e
	case *Ground:
		ds.Grounds[e.Id] = e
		ds.Elements[e.Id] = e
	case *GroundDisconnector:
		ds.GroundDisconnectors[e.Id] = e
		ds.Elements[e.Id] = e
	case *GroundingImpedance:
		ds.GroundingImpedances[e.Id] = e
		ds.Elements[e.Id] = e
	case *HydroGeneratingUnit:
		ds.HydroGeneratingUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *HydroPowerPlant:
		ds.HydroPowerPlants[e.Id] = e
		ds.Elements[e.Id] = e
	case *HydroPump:
		ds.HydroPumps[e.Id] = e
		ds.Elements[e.Id] = e
	case *Jumper:
		ds.Jumpers[e.Id] = e
		ds.Elements[e.Id] = e
	case *Junction:
		ds.Junctions[e.Id] = e
		ds.Elements[e.Id] = e
	case *Line:
		ds.Lines[e.Id] = e
		ds.Elements[e.Id] = e
	case *LinearShuntCompensator:
		ds.LinearShuntCompensators[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadAggregate:
		ds.LoadAggregates[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadArea:
		ds.LoadAreas[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadBreakSwitch:
		ds.LoadBreakSwitchs[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadComposite:
		ds.LoadComposites[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadGenericNonLinear:
		ds.LoadGenericNonLinears[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadMotor:
		ds.LoadMotors[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadResponseCharacteristic:
		ds.LoadResponseCharacteristics[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadStatic:
		ds.LoadStatics[e.Id] = e
		ds.Elements[e.Id] = e
	case *LoadUserDefined:
		ds.LoadUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *Location:
		ds.Locations[e.Id] = e
		ds.Elements[e.Id] = e
	case *MeasurementValueQuality:
		ds.MeasurementValueQualitys[e.Id] = e
		ds.Elements[e.Id] = e
	case *MeasurementValueSource:
		ds.MeasurementValueSources[e.Id] = e
		ds.Elements[e.Id] = e
	case *MechLoad1:
		ds.MechLoad1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *MechanicalLoadUserDefined:
		ds.MechanicalLoadUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *MutualCoupling:
		ds.MutualCouplings[e.Id] = e
		ds.Elements[e.Id] = e
	case *NonConformLoad:
		ds.NonConformLoads[e.Id] = e
		ds.Elements[e.Id] = e
	case *NonConformLoadGroup:
		ds.NonConformLoadGroups[e.Id] = e
		ds.Elements[e.Id] = e
	case *NonConformLoadSchedule:
		ds.NonConformLoadSchedules[e.Id] = e
		ds.Elements[e.Id] = e
	case *NonlinearShuntCompensator:
		ds.NonlinearShuntCompensators[e.Id] = e
		ds.Elements[e.Id] = e
	case *NonlinearShuntCompensatorPoint:
		ds.NonlinearShuntCompensatorPoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *NuclearGeneratingUnit:
		ds.NuclearGeneratingUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *OperationalLimitSet:
		ds.OperationalLimitSets[e.Id] = e
		ds.Elements[e.Id] = e
	case *OperationalLimitType:
		ds.OperationalLimitTypes[e.Id] = e
		ds.Elements[e.Id] = e
	case *OverexcLim2:
		ds.OverexcLim2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *OverexcLimIEEE:
		ds.OverexcLimIEEEs[e.Id] = e
		ds.Elements[e.Id] = e
	case *OverexcLimX1:
		ds.OverexcLimX1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *OverexcLimX2:
		ds.OverexcLimX2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *OverexcitationLimiterUserDefined:
		ds.OverexcitationLimiterUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *PFVArControllerType1UserDefined:
		ds.PFVArControllerType1UserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *PFVArControllerType2UserDefined:
		ds.PFVArControllerType2UserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *PFVArType1IEEEPFController:
		ds.PFVArType1IEEEPFControllers[e.Id] = e
		ds.Elements[e.Id] = e
	case *PFVArType1IEEEVArController:
		ds.PFVArType1IEEEVArControllers[e.Id] = e
		ds.Elements[e.Id] = e
	case *PFVArType2Common1:
		ds.PFVArType2Common1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *PFVArType2IEEEPFController:
		ds.PFVArType2IEEEPFControllers[e.Id] = e
		ds.Elements[e.Id] = e
	case *PFVArType2IEEEVArController:
		ds.PFVArType2IEEEVArControllers[e.Id] = e
		ds.Elements[e.Id] = e
	case *PetersenCoil:
		ds.PetersenCoils[e.Id] = e
		ds.Elements[e.Id] = e
	case *PhaseTapChangerAsymmetrical:
		ds.PhaseTapChangerAsymmetricals[e.Id] = e
		ds.Elements[e.Id] = e
	case *PhaseTapChangerLinear:
		ds.PhaseTapChangerLinears[e.Id] = e
		ds.Elements[e.Id] = e
	case *PhaseTapChangerSymmetrical:
		ds.PhaseTapChangerSymmetricals[e.Id] = e
		ds.Elements[e.Id] = e
	case *PhaseTapChangerTable:
		ds.PhaseTapChangerTables[e.Id] = e
		ds.Elements[e.Id] = e
	case *PhaseTapChangerTablePoint:
		ds.PhaseTapChangerTablePoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *PhaseTapChangerTabular:
		ds.PhaseTapChangerTabulars[e.Id] = e
		ds.Elements[e.Id] = e
	case *PhotoVoltaicUnit:
		ds.PhotoVoltaicUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *PositionPoint:
		ds.PositionPoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *PostLineSensor:
		ds.PostLineSensors[e.Id] = e
		ds.Elements[e.Id] = e
	case *PotentialTransformer:
		ds.PotentialTransformers[e.Id] = e
		ds.Elements[e.Id] = e
	case *PowerElectronicsConnection:
		ds.PowerElectronicsConnections[e.Id] = e
		ds.Elements[e.Id] = e
	case *PowerElectronicsWindUnit:
		ds.PowerElectronicsWindUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *PowerSystemStabilizerUserDefined:
		ds.PowerSystemStabilizerUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *PowerTransformer:
		ds.PowerTransformers[e.Id] = e
		ds.Elements[e.Id] = e
	case *PowerTransformerEnd:
		ds.PowerTransformerEnds[e.Id] = e
		ds.Elements[e.Id] = e
	case *ProprietaryParameterDynamics:
		ds.ProprietaryParameterDynamicss[e.Id] = e
		ds.Elements[e.Id] = e
	case *Pss1:
		ds.Pss1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *Pss1A:
		ds.Pss1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *Pss2B:
		ds.Pss2Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *Pss2ST:
		ds.Pss2STs[e.Id] = e
		ds.Elements[e.Id] = e
	case *Pss5:
		ds.Pss5s[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssELIN2:
		ds.PssELIN2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssIEEE1A:
		ds.PssIEEE1As[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssIEEE2B:
		ds.PssIEEE2Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssIEEE3B:
		ds.PssIEEE3Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssIEEE4B:
		ds.PssIEEE4Bs[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssPTIST1:
		ds.PssPTIST1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssPTIST3:
		ds.PssPTIST3s[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssRQB:
		ds.PssRQBs[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssSB4:
		ds.PssSB4s[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssSH:
		ds.PssSHs[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssSK:
		ds.PssSKs[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssSTAB2A:
		ds.PssSTAB2As[e.Id] = e
		ds.Elements[e.Id] = e
	case *PssWECC:
		ds.PssWECCs[e.Id] = e
		ds.Elements[e.Id] = e
	case *RaiseLowerCommand:
		ds.RaiseLowerCommands[e.Id] = e
		ds.Elements[e.Id] = e
	case *RatioTapChanger:
		ds.RatioTapChangers[e.Id] = e
		ds.Elements[e.Id] = e
	case *RatioTapChangerTable:
		ds.RatioTapChangerTables[e.Id] = e
		ds.Elements[e.Id] = e
	case *RatioTapChangerTablePoint:
		ds.RatioTapChangerTablePoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *ReactiveCapabilityCurve:
		ds.ReactiveCapabilityCurves[e.Id] = e
		ds.Elements[e.Id] = e
	case *RegularTimePoint:
		ds.RegularTimePoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *RegulatingControl:
		ds.RegulatingControls[e.Id] = e
		ds.Elements[e.Id] = e
	case *RegulationSchedule:
		ds.RegulationSchedules[e.Id] = e
		ds.Elements[e.Id] = e
	case *RemoteInputSignal:
		ds.RemoteInputSignals[e.Id] = e
		ds.Elements[e.Id] = e
	case *ReportingGroup:
		ds.ReportingGroups[e.Id] = e
		ds.Elements[e.Id] = e
	case *SVCUserDefined:
		ds.SVCUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *Season:
		ds.Seasons[e.Id] = e
		ds.Elements[e.Id] = e
	case *SeriesCompensator:
		ds.SeriesCompensators[e.Id] = e
		ds.Elements[e.Id] = e
	case *ServiceLocation:
		ds.ServiceLocations[e.Id] = e
		ds.Elements[e.Id] = e
	case *SetPoint:
		ds.SetPoints[e.Id] = e
		ds.Elements[e.Id] = e
	case *SolarGeneratingUnit:
		ds.SolarGeneratingUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *SolarPowerPlant:
		ds.SolarPowerPlants[e.Id] = e
		ds.Elements[e.Id] = e
	case *StaticVarCompensator:
		ds.StaticVarCompensators[e.Id] = e
		ds.Elements[e.Id] = e
	case *StationSupply:
		ds.StationSupplys[e.Id] = e
		ds.Elements[e.Id] = e
	case *StringMeasurement:
		ds.StringMeasurements[e.Id] = e
		ds.Elements[e.Id] = e
	case *StringMeasurementValue:
		ds.StringMeasurementValues[e.Id] = e
		ds.Elements[e.Id] = e
	case *SubGeographicalRegion:
		ds.SubGeographicalRegions[e.Id] = e
		ds.Elements[e.Id] = e
	case *SubLoadArea:
		ds.SubLoadAreas[e.Id] = e
		ds.Elements[e.Id] = e
	case *Substation:
		ds.Substations[e.Id] = e
		ds.Elements[e.Id] = e
	case *SurgeArrester:
		ds.SurgeArresters[e.Id] = e
		ds.Elements[e.Id] = e
	case *SvInjection:
		ds.SvInjections[e.Id] = e
		ds.Elements[e.Id] = e
	case *SvPowerFlow:
		ds.SvPowerFlows[e.Id] = e
		ds.Elements[e.Id] = e
	case *SvShuntCompensatorSections:
		ds.SvShuntCompensatorSectionss[e.Id] = e
		ds.Elements[e.Id] = e
	case *SvStatus:
		ds.SvStatuss[e.Id] = e
		ds.Elements[e.Id] = e
	case *SvSwitch:
		ds.SvSwitchs[e.Id] = e
		ds.Elements[e.Id] = e
	case *SvTapStep:
		ds.SvTapSteps[e.Id] = e
		ds.Elements[e.Id] = e
	case *SvVoltage:
		ds.SvVoltages[e.Id] = e
		ds.Elements[e.Id] = e
	case *Switch:
		ds.Switchs[e.Id] = e
		ds.Elements[e.Id] = e
	case *SwitchSchedule:
		ds.SwitchSchedules[e.Id] = e
		ds.Elements[e.Id] = e
	case *SynchronousMachine:
		ds.SynchronousMachines[e.Id] = e
		ds.Elements[e.Id] = e
	case *SynchronousMachineEquivalentCircuit:
		ds.SynchronousMachineEquivalentCircuits[e.Id] = e
		ds.Elements[e.Id] = e
	case *SynchronousMachineSimplified:
		ds.SynchronousMachineSimplifieds[e.Id] = e
		ds.Elements[e.Id] = e
	case *SynchronousMachineTimeConstantReactance:
		ds.SynchronousMachineTimeConstantReactances[e.Id] = e
		ds.Elements[e.Id] = e
	case *SynchronousMachineUserDefined:
		ds.SynchronousMachineUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *TapChangerControl:
		ds.TapChangerControls[e.Id] = e
		ds.Elements[e.Id] = e
	case *TapSchedule:
		ds.TapSchedules[e.Id] = e
		ds.Elements[e.Id] = e
	case *Terminal:
		ds.Terminals[e.Id] = e
		ds.Elements[e.Id] = e
	case *TextDiagramObject:
		ds.TextDiagramObjects[e.Id] = e
		ds.Elements[e.Id] = e
	case *ThermalGeneratingUnit:
		ds.ThermalGeneratingUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *TieFlow:
		ds.TieFlows[e.Id] = e
		ds.Elements[e.Id] = e
	case *TopologicalIsland:
		ds.TopologicalIslands[e.Id] = e
		ds.Elements[e.Id] = e
	case *TopologicalNode:
		ds.TopologicalNodes[e.Id] = e
		ds.Elements[e.Id] = e
	case *TurbLCFB1:
		ds.TurbLCFB1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *TurbineGovernorUserDefined:
		ds.TurbineGovernorUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *TurbineLoadControllerUserDefined:
		ds.TurbineLoadControllerUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *UnderexcLim2Simplified:
		ds.UnderexcLim2Simplifieds[e.Id] = e
		ds.Elements[e.Id] = e
	case *UnderexcLimIEEE1:
		ds.UnderexcLimIEEE1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *UnderexcLimIEEE2:
		ds.UnderexcLimIEEE2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *UnderexcLimX1:
		ds.UnderexcLimX1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *UnderexcLimX2:
		ds.UnderexcLimX2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *UnderexcitationLimiterUserDefined:
		ds.UnderexcitationLimiterUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *VAdjIEEE:
		ds.VAdjIEEEs[e.Id] = e
		ds.Elements[e.Id] = e
	case *VCompIEEEType1:
		ds.VCompIEEEType1s[e.Id] = e
		ds.Elements[e.Id] = e
	case *VCompIEEEType2:
		ds.VCompIEEEType2s[e.Id] = e
		ds.Elements[e.Id] = e
	case *VSCUserDefined:
		ds.VSCUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *ValueAliasSet:
		ds.ValueAliasSets[e.Id] = e
		ds.Elements[e.Id] = e
	case *ValueToAlias:
		ds.ValueToAliass[e.Id] = e
		ds.Elements[e.Id] = e
	case *VisibilityLayer:
		ds.VisibilityLayers[e.Id] = e
		ds.Elements[e.Id] = e
	case *VoltageAdjusterUserDefined:
		ds.VoltageAdjusterUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *VoltageCompensatorUserDefined:
		ds.VoltageCompensatorUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *VoltageLevel:
		ds.VoltageLevels[e.Id] = e
		ds.Elements[e.Id] = e
	case *VoltageLimit:
		ds.VoltageLimits[e.Id] = e
		ds.Elements[e.Id] = e
	case *VsCapabilityCurve:
		ds.VsCapabilityCurves[e.Id] = e
		ds.Elements[e.Id] = e
	case *VsConverter:
		ds.VsConverters[e.Id] = e
		ds.Elements[e.Id] = e
	case *WaveTrap:
		ds.WaveTraps[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindAeroConstIEC:
		ds.WindAeroConstIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindAeroOneDimIEC:
		ds.WindAeroOneDimIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindAeroTwoDimIEC:
		ds.WindAeroTwoDimIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContCurrLimIEC:
		ds.WindContCurrLimIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContPType3IEC:
		ds.WindContPType3IECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContPType4aIEC:
		ds.WindContPType4aIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContPType4bIEC:
		ds.WindContPType4bIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContPitchAngleIEC:
		ds.WindContPitchAngleIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContQIEC:
		ds.WindContQIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContQLimIEC:
		ds.WindContQLimIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContQPQULimIEC:
		ds.WindContQPQULimIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindContRotorRIEC:
		ds.WindContRotorRIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindDynamicsLookupTable:
		ds.WindDynamicsLookupTables[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindGenTurbineType1aIEC:
		ds.WindGenTurbineType1aIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindGenTurbineType1bIEC:
		ds.WindGenTurbineType1bIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindGenTurbineType2IEC:
		ds.WindGenTurbineType2IECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindGenType3aIEC:
		ds.WindGenType3aIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindGenType3bIEC:
		ds.WindGenType3bIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindGenType4IEC:
		ds.WindGenType4IECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindGeneratingUnit:
		ds.WindGeneratingUnits[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindMechIEC:
		ds.WindMechIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindPitchContPowerIEC:
		ds.WindPitchContPowerIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindPlantFreqPcontrolIEC:
		ds.WindPlantFreqPcontrolIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindPlantIEC:
		ds.WindPlantIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindPlantReactiveControlIEC:
		ds.WindPlantReactiveControlIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindPlantUserDefined:
		ds.WindPlantUserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindPowerPlant:
		ds.WindPowerPlants[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindProtectionIEC:
		ds.WindProtectionIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindRefFrameRotIEC:
		ds.WindRefFrameRotIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindTurbineType3IEC:
		ds.WindTurbineType3IECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindTurbineType4aIEC:
		ds.WindTurbineType4aIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindTurbineType4bIEC:
		ds.WindTurbineType4bIECs[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindType1or2UserDefined:
		ds.WindType1or2UserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	case *WindType3or4UserDefined:
		ds.WindType3or4UserDefineds[e.Id] = e
		ds.Elements[e.Id] = e
	default:
		fmt.Printf("Unknown type: %T\n", e)
	}
}

var StructMap = map[string]func() interface{}{
    "ACDCConverterDCTerminal": func() interface{} {
        return &ACDCConverterDCTerminal{}
    },
    "ACLineSegment": func() interface{} {
        return &ACLineSegment{}
    },
    "Accumulator": func() interface{} {
        return &Accumulator{}
    },
    "AccumulatorLimit": func() interface{} {
        return &AccumulatorLimit{}
    },
    "AccumulatorLimitSet": func() interface{} {
        return &AccumulatorLimitSet{}
    },
    "AccumulatorReset": func() interface{} {
        return &AccumulatorReset{}
    },
    "AccumulatorValue": func() interface{} {
        return &AccumulatorValue{}
    },
    "ActivePowerLimit": func() interface{} {
        return &ActivePowerLimit{}
    },
    "Analog": func() interface{} {
        return &Analog{}
    },
    "AnalogLimit": func() interface{} {
        return &AnalogLimit{}
    },
    "AnalogLimitSet": func() interface{} {
        return &AnalogLimitSet{}
    },
    "AnalogValue": func() interface{} {
        return &AnalogValue{}
    },
    "ApparentPowerLimit": func() interface{} {
        return &ApparentPowerLimit{}
    },
    "AsynchronousMachine": func() interface{} {
        return &AsynchronousMachine{}
    },
    "AsynchronousMachineEquivalentCircuit": func() interface{} {
        return &AsynchronousMachineEquivalentCircuit{}
    },
    "AsynchronousMachineTimeConstantReactance": func() interface{} {
        return &AsynchronousMachineTimeConstantReactance{}
    },
    "AsynchronousMachineUserDefined": func() interface{} {
        return &AsynchronousMachineUserDefined{}
    },
    "BaseVoltage": func() interface{} {
        return &BaseVoltage{}
    },
    "BatteryUnit": func() interface{} {
        return &BatteryUnit{}
    },
    "Bay": func() interface{} {
        return &Bay{}
    },
    "BoundaryPoint": func() interface{} {
        return &BoundaryPoint{}
    },
    "Breaker": func() interface{} {
        return &Breaker{}
    },
    "BusNameMarker": func() interface{} {
        return &BusNameMarker{}
    },
    "BusbarSection": func() interface{} {
        return &BusbarSection{}
    },
    "CAESPlant": func() interface{} {
        return &CAESPlant{}
    },
    "CSCUserDefined": func() interface{} {
        return &CSCUserDefined{}
    },
    "Clamp": func() interface{} {
        return &Clamp{}
    },
    "CogenerationPlant": func() interface{} {
        return &CogenerationPlant{}
    },
    "CombinedCyclePlant": func() interface{} {
        return &CombinedCyclePlant{}
    },
    "Command": func() interface{} {
        return &Command{}
    },
    "ConformLoad": func() interface{} {
        return &ConformLoad{}
    },
    "ConformLoadGroup": func() interface{} {
        return &ConformLoadGroup{}
    },
    "ConformLoadSchedule": func() interface{} {
        return &ConformLoadSchedule{}
    },
    "ConnectivityNode": func() interface{} {
        return &ConnectivityNode{}
    },
    "ControlArea": func() interface{} {
        return &ControlArea{}
    },
    "ControlAreaGeneratingUnit": func() interface{} {
        return &ControlAreaGeneratingUnit{}
    },
    "CoordinateSystem": func() interface{} {
        return &CoordinateSystem{}
    },
    "CsConverter": func() interface{} {
        return &CsConverter{}
    },
    "CurrentLimit": func() interface{} {
        return &CurrentLimit{}
    },
    "CurrentTransformer": func() interface{} {
        return &CurrentTransformer{}
    },
    "CurveData": func() interface{} {
        return &CurveData{}
    },
    "Cut": func() interface{} {
        return &Cut{}
    },
    "DCBreaker": func() interface{} {
        return &DCBreaker{}
    },
    "DCBusbar": func() interface{} {
        return &DCBusbar{}
    },
    "DCChopper": func() interface{} {
        return &DCChopper{}
    },
    "DCConverterUnit": func() interface{} {
        return &DCConverterUnit{}
    },
    "DCDisconnector": func() interface{} {
        return &DCDisconnector{}
    },
    "DCGround": func() interface{} {
        return &DCGround{}
    },
    "DCLine": func() interface{} {
        return &DCLine{}
    },
    "DCLineSegment": func() interface{} {
        return &DCLineSegment{}
    },
    "DCNode": func() interface{} {
        return &DCNode{}
    },
    "DCSeriesDevice": func() interface{} {
        return &DCSeriesDevice{}
    },
    "DCShunt": func() interface{} {
        return &DCShunt{}
    },
    "DCSwitch": func() interface{} {
        return &DCSwitch{}
    },
    "DCTerminal": func() interface{} {
        return &DCTerminal{}
    },
    "DCTopologicalIsland": func() interface{} {
        return &DCTopologicalIsland{}
    },
    "DCTopologicalNode": func() interface{} {
        return &DCTopologicalNode{}
    },
    "DayType": func() interface{} {
        return &DayType{}
    },
    "Diagram": func() interface{} {
        return &Diagram{}
    },
    "DiagramObject": func() interface{} {
        return &DiagramObject{}
    },
    "DiagramObjectGluePoint": func() interface{} {
        return &DiagramObjectGluePoint{}
    },
    "DiagramObjectPoint": func() interface{} {
        return &DiagramObjectPoint{}
    },
    "DiagramObjectStyle": func() interface{} {
        return &DiagramObjectStyle{}
    },
    "DiagramStyle": func() interface{} {
        return &DiagramStyle{}
    },
    "DifferenceModel": func() interface{} {
        return &DifferenceModel{}
    },
    "DiscExcContIEEEDEC1A": func() interface{} {
        return &DiscExcContIEEEDEC1A{}
    },
    "DiscExcContIEEEDEC2A": func() interface{} {
        return &DiscExcContIEEEDEC2A{}
    },
    "DiscExcContIEEEDEC3A": func() interface{} {
        return &DiscExcContIEEEDEC3A{}
    },
    "DisconnectingCircuitBreaker": func() interface{} {
        return &DisconnectingCircuitBreaker{}
    },
    "Disconnector": func() interface{} {
        return &Disconnector{}
    },
    "DiscontinuousExcitationControlUserDefined": func() interface{} {
        return &DiscontinuousExcitationControlUserDefined{}
    },
    "Discrete": func() interface{} {
        return &Discrete{}
    },
    "DiscreteValue": func() interface{} {
        return &DiscreteValue{}
    },
    "EnergyConsumer": func() interface{} {
        return &EnergyConsumer{}
    },
    "EnergySchedulingType": func() interface{} {
        return &EnergySchedulingType{}
    },
    "EnergySource": func() interface{} {
        return &EnergySource{}
    },
    "Equipment": func() interface{} {
        return &Equipment{}
    },
    "EquivalentBranch": func() interface{} {
        return &EquivalentBranch{}
    },
    "EquivalentInjection": func() interface{} {
        return &EquivalentInjection{}
    },
    "EquivalentNetwork": func() interface{} {
        return &EquivalentNetwork{}
    },
    "EquivalentShunt": func() interface{} {
        return &EquivalentShunt{}
    },
    "ExcAC1A": func() interface{} {
        return &ExcAC1A{}
    },
    "ExcAC2A": func() interface{} {
        return &ExcAC2A{}
    },
    "ExcAC3A": func() interface{} {
        return &ExcAC3A{}
    },
    "ExcAC4A": func() interface{} {
        return &ExcAC4A{}
    },
    "ExcAC5A": func() interface{} {
        return &ExcAC5A{}
    },
    "ExcAC6A": func() interface{} {
        return &ExcAC6A{}
    },
    "ExcAC8B": func() interface{} {
        return &ExcAC8B{}
    },
    "ExcANS": func() interface{} {
        return &ExcANS{}
    },
    "ExcAVR1": func() interface{} {
        return &ExcAVR1{}
    },
    "ExcAVR2": func() interface{} {
        return &ExcAVR2{}
    },
    "ExcAVR3": func() interface{} {
        return &ExcAVR3{}
    },
    "ExcAVR4": func() interface{} {
        return &ExcAVR4{}
    },
    "ExcAVR5": func() interface{} {
        return &ExcAVR5{}
    },
    "ExcAVR7": func() interface{} {
        return &ExcAVR7{}
    },
    "ExcBBC": func() interface{} {
        return &ExcBBC{}
    },
    "ExcCZ": func() interface{} {
        return &ExcCZ{}
    },
    "ExcDC1A": func() interface{} {
        return &ExcDC1A{}
    },
    "ExcDC2A": func() interface{} {
        return &ExcDC2A{}
    },
    "ExcDC3A": func() interface{} {
        return &ExcDC3A{}
    },
    "ExcDC3A1": func() interface{} {
        return &ExcDC3A1{}
    },
    "ExcELIN1": func() interface{} {
        return &ExcELIN1{}
    },
    "ExcELIN2": func() interface{} {
        return &ExcELIN2{}
    },
    "ExcHU": func() interface{} {
        return &ExcHU{}
    },
    "ExcIEEEAC1A": func() interface{} {
        return &ExcIEEEAC1A{}
    },
    "ExcIEEEAC2A": func() interface{} {
        return &ExcIEEEAC2A{}
    },
    "ExcIEEEAC3A": func() interface{} {
        return &ExcIEEEAC3A{}
    },
    "ExcIEEEAC4A": func() interface{} {
        return &ExcIEEEAC4A{}
    },
    "ExcIEEEAC5A": func() interface{} {
        return &ExcIEEEAC5A{}
    },
    "ExcIEEEAC6A": func() interface{} {
        return &ExcIEEEAC6A{}
    },
    "ExcIEEEAC7B": func() interface{} {
        return &ExcIEEEAC7B{}
    },
    "ExcIEEEAC8B": func() interface{} {
        return &ExcIEEEAC8B{}
    },
    "ExcIEEEDC1A": func() interface{} {
        return &ExcIEEEDC1A{}
    },
    "ExcIEEEDC2A": func() interface{} {
        return &ExcIEEEDC2A{}
    },
    "ExcIEEEDC3A": func() interface{} {
        return &ExcIEEEDC3A{}
    },
    "ExcIEEEDC4B": func() interface{} {
        return &ExcIEEEDC4B{}
    },
    "ExcIEEEST1A": func() interface{} {
        return &ExcIEEEST1A{}
    },
    "ExcIEEEST2A": func() interface{} {
        return &ExcIEEEST2A{}
    },
    "ExcIEEEST3A": func() interface{} {
        return &ExcIEEEST3A{}
    },
    "ExcIEEEST4B": func() interface{} {
        return &ExcIEEEST4B{}
    },
    "ExcIEEEST5B": func() interface{} {
        return &ExcIEEEST5B{}
    },
    "ExcIEEEST6B": func() interface{} {
        return &ExcIEEEST6B{}
    },
    "ExcIEEEST7B": func() interface{} {
        return &ExcIEEEST7B{}
    },
    "ExcNI": func() interface{} {
        return &ExcNI{}
    },
    "ExcOEX3T": func() interface{} {
        return &ExcOEX3T{}
    },
    "ExcPIC": func() interface{} {
        return &ExcPIC{}
    },
    "ExcREXS": func() interface{} {
        return &ExcREXS{}
    },
    "ExcRQB": func() interface{} {
        return &ExcRQB{}
    },
    "ExcSCRX": func() interface{} {
        return &ExcSCRX{}
    },
    "ExcSEXS": func() interface{} {
        return &ExcSEXS{}
    },
    "ExcSK": func() interface{} {
        return &ExcSK{}
    },
    "ExcST1A": func() interface{} {
        return &ExcST1A{}
    },
    "ExcST2A": func() interface{} {
        return &ExcST2A{}
    },
    "ExcST3A": func() interface{} {
        return &ExcST3A{}
    },
    "ExcST4B": func() interface{} {
        return &ExcST4B{}
    },
    "ExcST6B": func() interface{} {
        return &ExcST6B{}
    },
    "ExcST7B": func() interface{} {
        return &ExcST7B{}
    },
    "ExcitationSystemUserDefined": func() interface{} {
        return &ExcitationSystemUserDefined{}
    },
    "ExternalNetworkInjection": func() interface{} {
        return &ExternalNetworkInjection{}
    },
    "FaultIndicator": func() interface{} {
        return &FaultIndicator{}
    },
    "FossilFuel": func() interface{} {
        return &FossilFuel{}
    },
    "FullModel": func() interface{} {
        return &FullModel{}
    },
    "Fuse": func() interface{} {
        return &Fuse{}
    },
    "GenICompensationForGenJ": func() interface{} {
        return &GenICompensationForGenJ{}
    },
    "GeneratingUnit": func() interface{} {
        return &GeneratingUnit{}
    },
    "GeographicalRegion": func() interface{} {
        return &GeographicalRegion{}
    },
    "GovCT1": func() interface{} {
        return &GovCT1{}
    },
    "GovCT2": func() interface{} {
        return &GovCT2{}
    },
    "GovGAST": func() interface{} {
        return &GovGAST{}
    },
    "GovGAST1": func() interface{} {
        return &GovGAST1{}
    },
    "GovGAST2": func() interface{} {
        return &GovGAST2{}
    },
    "GovGAST3": func() interface{} {
        return &GovGAST3{}
    },
    "GovGAST4": func() interface{} {
        return &GovGAST4{}
    },
    "GovGASTWD": func() interface{} {
        return &GovGASTWD{}
    },
    "GovHydro1": func() interface{} {
        return &GovHydro1{}
    },
    "GovHydro2": func() interface{} {
        return &GovHydro2{}
    },
    "GovHydro3": func() interface{} {
        return &GovHydro3{}
    },
    "GovHydro4": func() interface{} {
        return &GovHydro4{}
    },
    "GovHydroDD": func() interface{} {
        return &GovHydroDD{}
    },
    "GovHydroFrancis": func() interface{} {
        return &GovHydroFrancis{}
    },
    "GovHydroIEEE0": func() interface{} {
        return &GovHydroIEEE0{}
    },
    "GovHydroIEEE2": func() interface{} {
        return &GovHydroIEEE2{}
    },
    "GovHydroPID": func() interface{} {
        return &GovHydroPID{}
    },
    "GovHydroPID2": func() interface{} {
        return &GovHydroPID2{}
    },
    "GovHydroPelton": func() interface{} {
        return &GovHydroPelton{}
    },
    "GovHydroR": func() interface{} {
        return &GovHydroR{}
    },
    "GovHydroWEH": func() interface{} {
        return &GovHydroWEH{}
    },
    "GovHydroWPID": func() interface{} {
        return &GovHydroWPID{}
    },
    "GovSteam0": func() interface{} {
        return &GovSteam0{}
    },
    "GovSteam1": func() interface{} {
        return &GovSteam1{}
    },
    "GovSteam2": func() interface{} {
        return &GovSteam2{}
    },
    "GovSteamBB": func() interface{} {
        return &GovSteamBB{}
    },
    "GovSteamCC": func() interface{} {
        return &GovSteamCC{}
    },
    "GovSteamEU": func() interface{} {
        return &GovSteamEU{}
    },
    "GovSteamFV2": func() interface{} {
        return &GovSteamFV2{}
    },
    "GovSteamFV3": func() interface{} {
        return &GovSteamFV3{}
    },
    "GovSteamFV4": func() interface{} {
        return &GovSteamFV4{}
    },
    "GovSteamIEEE1": func() interface{} {
        return &GovSteamIEEE1{}
    },
    "GovSteamSGO": func() interface{} {
        return &GovSteamSGO{}
    },
    "GrossToNetActivePowerCurve": func() interface{} {
        return &GrossToNetActivePowerCurve{}
    },
    "Ground": func() interface{} {
        return &Ground{}
    },
    "GroundDisconnector": func() interface{} {
        return &GroundDisconnector{}
    },
    "GroundingImpedance": func() interface{} {
        return &GroundingImpedance{}
    },
    "HydroGeneratingUnit": func() interface{} {
        return &HydroGeneratingUnit{}
    },
    "HydroPowerPlant": func() interface{} {
        return &HydroPowerPlant{}
    },
    "HydroPump": func() interface{} {
        return &HydroPump{}
    },
    "Jumper": func() interface{} {
        return &Jumper{}
    },
    "Junction": func() interface{} {
        return &Junction{}
    },
    "Line": func() interface{} {
        return &Line{}
    },
    "LinearShuntCompensator": func() interface{} {
        return &LinearShuntCompensator{}
    },
    "LoadAggregate": func() interface{} {
        return &LoadAggregate{}
    },
    "LoadArea": func() interface{} {
        return &LoadArea{}
    },
    "LoadBreakSwitch": func() interface{} {
        return &LoadBreakSwitch{}
    },
    "LoadComposite": func() interface{} {
        return &LoadComposite{}
    },
    "LoadGenericNonLinear": func() interface{} {
        return &LoadGenericNonLinear{}
    },
    "LoadMotor": func() interface{} {
        return &LoadMotor{}
    },
    "LoadResponseCharacteristic": func() interface{} {
        return &LoadResponseCharacteristic{}
    },
    "LoadStatic": func() interface{} {
        return &LoadStatic{}
    },
    "LoadUserDefined": func() interface{} {
        return &LoadUserDefined{}
    },
    "Location": func() interface{} {
        return &Location{}
    },
    "MeasurementValueQuality": func() interface{} {
        return &MeasurementValueQuality{}
    },
    "MeasurementValueSource": func() interface{} {
        return &MeasurementValueSource{}
    },
    "MechLoad1": func() interface{} {
        return &MechLoad1{}
    },
    "MechanicalLoadUserDefined": func() interface{} {
        return &MechanicalLoadUserDefined{}
    },
    "MutualCoupling": func() interface{} {
        return &MutualCoupling{}
    },
    "NonConformLoad": func() interface{} {
        return &NonConformLoad{}
    },
    "NonConformLoadGroup": func() interface{} {
        return &NonConformLoadGroup{}
    },
    "NonConformLoadSchedule": func() interface{} {
        return &NonConformLoadSchedule{}
    },
    "NonlinearShuntCompensator": func() interface{} {
        return &NonlinearShuntCompensator{}
    },
    "NonlinearShuntCompensatorPoint": func() interface{} {
        return &NonlinearShuntCompensatorPoint{}
    },
    "NuclearGeneratingUnit": func() interface{} {
        return &NuclearGeneratingUnit{}
    },
    "OperationalLimitSet": func() interface{} {
        return &OperationalLimitSet{}
    },
    "OperationalLimitType": func() interface{} {
        return &OperationalLimitType{}
    },
    "OverexcLim2": func() interface{} {
        return &OverexcLim2{}
    },
    "OverexcLimIEEE": func() interface{} {
        return &OverexcLimIEEE{}
    },
    "OverexcLimX1": func() interface{} {
        return &OverexcLimX1{}
    },
    "OverexcLimX2": func() interface{} {
        return &OverexcLimX2{}
    },
    "OverexcitationLimiterUserDefined": func() interface{} {
        return &OverexcitationLimiterUserDefined{}
    },
    "PFVArControllerType1UserDefined": func() interface{} {
        return &PFVArControllerType1UserDefined{}
    },
    "PFVArControllerType2UserDefined": func() interface{} {
        return &PFVArControllerType2UserDefined{}
    },
    "PFVArType1IEEEPFController": func() interface{} {
        return &PFVArType1IEEEPFController{}
    },
    "PFVArType1IEEEVArController": func() interface{} {
        return &PFVArType1IEEEVArController{}
    },
    "PFVArType2Common1": func() interface{} {
        return &PFVArType2Common1{}
    },
    "PFVArType2IEEEPFController": func() interface{} {
        return &PFVArType2IEEEPFController{}
    },
    "PFVArType2IEEEVArController": func() interface{} {
        return &PFVArType2IEEEVArController{}
    },
    "PetersenCoil": func() interface{} {
        return &PetersenCoil{}
    },
    "PhaseTapChangerAsymmetrical": func() interface{} {
        return &PhaseTapChangerAsymmetrical{}
    },
    "PhaseTapChangerLinear": func() interface{} {
        return &PhaseTapChangerLinear{}
    },
    "PhaseTapChangerSymmetrical": func() interface{} {
        return &PhaseTapChangerSymmetrical{}
    },
    "PhaseTapChangerTable": func() interface{} {
        return &PhaseTapChangerTable{}
    },
    "PhaseTapChangerTablePoint": func() interface{} {
        return &PhaseTapChangerTablePoint{}
    },
    "PhaseTapChangerTabular": func() interface{} {
        return &PhaseTapChangerTabular{}
    },
    "PhotoVoltaicUnit": func() interface{} {
        return &PhotoVoltaicUnit{}
    },
    "PositionPoint": func() interface{} {
        return &PositionPoint{}
    },
    "PostLineSensor": func() interface{} {
        return &PostLineSensor{}
    },
    "PotentialTransformer": func() interface{} {
        return &PotentialTransformer{}
    },
    "PowerElectronicsConnection": func() interface{} {
        return &PowerElectronicsConnection{}
    },
    "PowerElectronicsWindUnit": func() interface{} {
        return &PowerElectronicsWindUnit{}
    },
    "PowerSystemStabilizerUserDefined": func() interface{} {
        return &PowerSystemStabilizerUserDefined{}
    },
    "PowerTransformer": func() interface{} {
        return &PowerTransformer{}
    },
    "PowerTransformerEnd": func() interface{} {
        return &PowerTransformerEnd{}
    },
    "ProprietaryParameterDynamics": func() interface{} {
        return &ProprietaryParameterDynamics{}
    },
    "Pss1": func() interface{} {
        return &Pss1{}
    },
    "Pss1A": func() interface{} {
        return &Pss1A{}
    },
    "Pss2B": func() interface{} {
        return &Pss2B{}
    },
    "Pss2ST": func() interface{} {
        return &Pss2ST{}
    },
    "Pss5": func() interface{} {
        return &Pss5{}
    },
    "PssELIN2": func() interface{} {
        return &PssELIN2{}
    },
    "PssIEEE1A": func() interface{} {
        return &PssIEEE1A{}
    },
    "PssIEEE2B": func() interface{} {
        return &PssIEEE2B{}
    },
    "PssIEEE3B": func() interface{} {
        return &PssIEEE3B{}
    },
    "PssIEEE4B": func() interface{} {
        return &PssIEEE4B{}
    },
    "PssPTIST1": func() interface{} {
        return &PssPTIST1{}
    },
    "PssPTIST3": func() interface{} {
        return &PssPTIST3{}
    },
    "PssRQB": func() interface{} {
        return &PssRQB{}
    },
    "PssSB4": func() interface{} {
        return &PssSB4{}
    },
    "PssSH": func() interface{} {
        return &PssSH{}
    },
    "PssSK": func() interface{} {
        return &PssSK{}
    },
    "PssSTAB2A": func() interface{} {
        return &PssSTAB2A{}
    },
    "PssWECC": func() interface{} {
        return &PssWECC{}
    },
    "RaiseLowerCommand": func() interface{} {
        return &RaiseLowerCommand{}
    },
    "RatioTapChanger": func() interface{} {
        return &RatioTapChanger{}
    },
    "RatioTapChangerTable": func() interface{} {
        return &RatioTapChangerTable{}
    },
    "RatioTapChangerTablePoint": func() interface{} {
        return &RatioTapChangerTablePoint{}
    },
    "ReactiveCapabilityCurve": func() interface{} {
        return &ReactiveCapabilityCurve{}
    },
    "RegularTimePoint": func() interface{} {
        return &RegularTimePoint{}
    },
    "RegulatingControl": func() interface{} {
        return &RegulatingControl{}
    },
    "RegulationSchedule": func() interface{} {
        return &RegulationSchedule{}
    },
    "RemoteInputSignal": func() interface{} {
        return &RemoteInputSignal{}
    },
    "ReportingGroup": func() interface{} {
        return &ReportingGroup{}
    },
    "SVCUserDefined": func() interface{} {
        return &SVCUserDefined{}
    },
    "Season": func() interface{} {
        return &Season{}
    },
    "SeriesCompensator": func() interface{} {
        return &SeriesCompensator{}
    },
    "ServiceLocation": func() interface{} {
        return &ServiceLocation{}
    },
    "SetPoint": func() interface{} {
        return &SetPoint{}
    },
    "SolarGeneratingUnit": func() interface{} {
        return &SolarGeneratingUnit{}
    },
    "SolarPowerPlant": func() interface{} {
        return &SolarPowerPlant{}
    },
    "StaticVarCompensator": func() interface{} {
        return &StaticVarCompensator{}
    },
    "StationSupply": func() interface{} {
        return &StationSupply{}
    },
    "StringMeasurement": func() interface{} {
        return &StringMeasurement{}
    },
    "StringMeasurementValue": func() interface{} {
        return &StringMeasurementValue{}
    },
    "SubGeographicalRegion": func() interface{} {
        return &SubGeographicalRegion{}
    },
    "SubLoadArea": func() interface{} {
        return &SubLoadArea{}
    },
    "Substation": func() interface{} {
        return &Substation{}
    },
    "SurgeArrester": func() interface{} {
        return &SurgeArrester{}
    },
    "SvInjection": func() interface{} {
        return &SvInjection{}
    },
    "SvPowerFlow": func() interface{} {
        return &SvPowerFlow{}
    },
    "SvShuntCompensatorSections": func() interface{} {
        return &SvShuntCompensatorSections{}
    },
    "SvStatus": func() interface{} {
        return &SvStatus{}
    },
    "SvSwitch": func() interface{} {
        return &SvSwitch{}
    },
    "SvTapStep": func() interface{} {
        return &SvTapStep{}
    },
    "SvVoltage": func() interface{} {
        return &SvVoltage{}
    },
    "Switch": func() interface{} {
        return &Switch{}
    },
    "SwitchSchedule": func() interface{} {
        return &SwitchSchedule{}
    },
    "SynchronousMachine": func() interface{} {
        return &SynchronousMachine{}
    },
    "SynchronousMachineEquivalentCircuit": func() interface{} {
        return &SynchronousMachineEquivalentCircuit{}
    },
    "SynchronousMachineSimplified": func() interface{} {
        return &SynchronousMachineSimplified{}
    },
    "SynchronousMachineTimeConstantReactance": func() interface{} {
        return &SynchronousMachineTimeConstantReactance{}
    },
    "SynchronousMachineUserDefined": func() interface{} {
        return &SynchronousMachineUserDefined{}
    },
    "TapChangerControl": func() interface{} {
        return &TapChangerControl{}
    },
    "TapSchedule": func() interface{} {
        return &TapSchedule{}
    },
    "Terminal": func() interface{} {
        return &Terminal{}
    },
    "TextDiagramObject": func() interface{} {
        return &TextDiagramObject{}
    },
    "ThermalGeneratingUnit": func() interface{} {
        return &ThermalGeneratingUnit{}
    },
    "TieFlow": func() interface{} {
        return &TieFlow{}
    },
    "TopologicalIsland": func() interface{} {
        return &TopologicalIsland{}
    },
    "TopologicalNode": func() interface{} {
        return &TopologicalNode{}
    },
    "TurbLCFB1": func() interface{} {
        return &TurbLCFB1{}
    },
    "TurbineGovernorUserDefined": func() interface{} {
        return &TurbineGovernorUserDefined{}
    },
    "TurbineLoadControllerUserDefined": func() interface{} {
        return &TurbineLoadControllerUserDefined{}
    },
    "UnderexcLim2Simplified": func() interface{} {
        return &UnderexcLim2Simplified{}
    },
    "UnderexcLimIEEE1": func() interface{} {
        return &UnderexcLimIEEE1{}
    },
    "UnderexcLimIEEE2": func() interface{} {
        return &UnderexcLimIEEE2{}
    },
    "UnderexcLimX1": func() interface{} {
        return &UnderexcLimX1{}
    },
    "UnderexcLimX2": func() interface{} {
        return &UnderexcLimX2{}
    },
    "UnderexcitationLimiterUserDefined": func() interface{} {
        return &UnderexcitationLimiterUserDefined{}
    },
    "VAdjIEEE": func() interface{} {
        return &VAdjIEEE{}
    },
    "VCompIEEEType1": func() interface{} {
        return &VCompIEEEType1{}
    },
    "VCompIEEEType2": func() interface{} {
        return &VCompIEEEType2{}
    },
    "VSCUserDefined": func() interface{} {
        return &VSCUserDefined{}
    },
    "ValueAliasSet": func() interface{} {
        return &ValueAliasSet{}
    },
    "ValueToAlias": func() interface{} {
        return &ValueToAlias{}
    },
    "VisibilityLayer": func() interface{} {
        return &VisibilityLayer{}
    },
    "VoltageAdjusterUserDefined": func() interface{} {
        return &VoltageAdjusterUserDefined{}
    },
    "VoltageCompensatorUserDefined": func() interface{} {
        return &VoltageCompensatorUserDefined{}
    },
    "VoltageLevel": func() interface{} {
        return &VoltageLevel{}
    },
    "VoltageLimit": func() interface{} {
        return &VoltageLimit{}
    },
    "VsCapabilityCurve": func() interface{} {
        return &VsCapabilityCurve{}
    },
    "VsConverter": func() interface{} {
        return &VsConverter{}
    },
    "WaveTrap": func() interface{} {
        return &WaveTrap{}
    },
    "WindAeroConstIEC": func() interface{} {
        return &WindAeroConstIEC{}
    },
    "WindAeroOneDimIEC": func() interface{} {
        return &WindAeroOneDimIEC{}
    },
    "WindAeroTwoDimIEC": func() interface{} {
        return &WindAeroTwoDimIEC{}
    },
    "WindContCurrLimIEC": func() interface{} {
        return &WindContCurrLimIEC{}
    },
    "WindContPType3IEC": func() interface{} {
        return &WindContPType3IEC{}
    },
    "WindContPType4aIEC": func() interface{} {
        return &WindContPType4aIEC{}
    },
    "WindContPType4bIEC": func() interface{} {
        return &WindContPType4bIEC{}
    },
    "WindContPitchAngleIEC": func() interface{} {
        return &WindContPitchAngleIEC{}
    },
    "WindContQIEC": func() interface{} {
        return &WindContQIEC{}
    },
    "WindContQLimIEC": func() interface{} {
        return &WindContQLimIEC{}
    },
    "WindContQPQULimIEC": func() interface{} {
        return &WindContQPQULimIEC{}
    },
    "WindContRotorRIEC": func() interface{} {
        return &WindContRotorRIEC{}
    },
    "WindDynamicsLookupTable": func() interface{} {
        return &WindDynamicsLookupTable{}
    },
    "WindGenTurbineType1aIEC": func() interface{} {
        return &WindGenTurbineType1aIEC{}
    },
    "WindGenTurbineType1bIEC": func() interface{} {
        return &WindGenTurbineType1bIEC{}
    },
    "WindGenTurbineType2IEC": func() interface{} {
        return &WindGenTurbineType2IEC{}
    },
    "WindGenType3aIEC": func() interface{} {
        return &WindGenType3aIEC{}
    },
    "WindGenType3bIEC": func() interface{} {
        return &WindGenType3bIEC{}
    },
    "WindGenType4IEC": func() interface{} {
        return &WindGenType4IEC{}
    },
    "WindGeneratingUnit": func() interface{} {
        return &WindGeneratingUnit{}
    },
    "WindMechIEC": func() interface{} {
        return &WindMechIEC{}
    },
    "WindPitchContPowerIEC": func() interface{} {
        return &WindPitchContPowerIEC{}
    },
    "WindPlantFreqPcontrolIEC": func() interface{} {
        return &WindPlantFreqPcontrolIEC{}
    },
    "WindPlantIEC": func() interface{} {
        return &WindPlantIEC{}
    },
    "WindPlantReactiveControlIEC": func() interface{} {
        return &WindPlantReactiveControlIEC{}
    },
    "WindPlantUserDefined": func() interface{} {
        return &WindPlantUserDefined{}
    },
    "WindPowerPlant": func() interface{} {
        return &WindPowerPlant{}
    },
    "WindProtectionIEC": func() interface{} {
        return &WindProtectionIEC{}
    },
    "WindRefFrameRotIEC": func() interface{} {
        return &WindRefFrameRotIEC{}
    },
    "WindTurbineType3IEC": func() interface{} {
        return &WindTurbineType3IEC{}
    },
    "WindTurbineType4aIEC": func() interface{} {
        return &WindTurbineType4aIEC{}
    },
    "WindTurbineType4bIEC": func() interface{} {
        return &WindTurbineType4bIEC{}
    },
    "WindType1or2UserDefined": func() interface{} {
        return &WindType1or2UserDefined{}
    },
    "WindType3or4UserDefined": func() interface{} {
        return &WindType3or4UserDefined{}
    },
}
